// Sync a YAML map of username:github-user to cloud-init user data

package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
	"k8s.io/klog/v2"
)

var (
	userMapPath     = flag.String("usermap", "", "path to usermap")
	groupMapPath    = flag.String("groupmap", "", "path to groupmap")
	nodeMapPath     = flag.String("nodemap", "", "path to nodemap")
	outPath         = flag.String("out", "./out", "path to output directory")
	nodesFlag       = flag.String("nodes", "", "comma-delimited list of nodes to setup (default: all)")
	defaultGroup    = flag.String("default-group", "cafe", "default group to add users to")
	defaultShell    = flag.String("default-shell", "fish", "default shell")
	orgGroups       = flag.Bool("create-org-groups", false, "create groups for GitHub organizations")
	startUID        = flag.Int("uid", 2001, "UID to start users at")
	gitHubTokenFile = flag.String("github-token-file", "", "github token secret file")
)

type userConfig struct {
	Name       string   `yaml:"name"`
	GitHub     string   `yaml:"github"`
	LoginGroup string   `yaml:"login_group"`
	Groups     []string `yaml:"groups"`
	Shell      string   `yaml:"shell"`
	Exclude    []string `yaml:"exclude"`
}

type userMap struct {
	Users []userConfig `yaml:"users"`
}

type groupMap struct {
	Groups map[string]int64 `yaml:"groups"`
}

type node struct {
	Name         string
	Arch         string
	OS           string
	Distro       string
	ExcludeUsers []string `yaml:"exclude_users"`
}

type nodeMap struct {
	Nodes []node `yaml:"nodes"`
}

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	f, err := os.Open(*userMapPath)
	if err != nil {
		klog.Exitf("open %s: %v", *userMapPath, err)
	}

	um, err := loadUserMap(f)
	if err != nil {
		klog.Exitf("load from %s: %v", *userMapPath, err)
	}

	g, err := os.Open(*groupMapPath)
	if err != nil {
		klog.Exitf("open %s: %v", *groupMapPath, err)
	}

	gm, err := loadGroupMap(g)
	if err != nil {
		klog.Exitf("load from %s: %v", *groupMapPath, err)
	}

	nf, err := os.Open(*nodeMapPath)
	if err != nil {
		klog.Exitf("open %s: %v", *nodeMapPath, err)
	}
	nm, err := loadNodeMap(nf)
	if err != nil {
		klog.Exitf("load from %s: %v", *nodeMapPath, err)
	}

	onlyNodes := map[string]bool{}
	for _, no := range strings.Split(*nodesFlag, ",") {
		if no != "" {
			onlyNodes[no] = true
		}
	}

	for _, n := range nm.Nodes {
		if len(onlyNodes) > 0 && !onlyNodes[n.Name] {
			klog.Infof("skipping %s - not in %s", n.Name, onlyNodes)
			continue
		}
		path := filepath.Join(*outPath, fmt.Sprintf("%s.yaml", n.Name))
		klog.Infof("node %s -> %s", n.Name, path)
		if err := dumpPlaybook(um, gm, n, path); err != nil {
			klog.Exitf("%s: %s", path, err)
		}
	}
}

func dumpPlaybook(um *userMap, gm *groupMap, n node, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create: %v", err)
	}

	if _, err := fmt.Fprintf(f, "# autogenerated by:\n# generate-ansible %s\n---\n", strings.Join(os.Args[1:], " ")); err != nil {
		return fmt.Errorf("write failed to %s: %v", path, err)
	}

	pbs := []playbook{}
	pbs = append(pbs, createPlaybook(um, gm, n))

	bs, err := yaml.Marshal(pbs)
	if err != nil {
		return fmt.Errorf("marshal err: %v", err)
	}
	if _, err := f.Write(bs); err != nil {
		return fmt.Errorf("write: %v", err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("close %s: %v", path, err)
	}
	return nil
}

type ansibleGroup struct {
	Name   string `yaml:"name"`
	ID     int64  `yaml:"gid"`
	State  string `yaml:"state"`
	System bool   `yaml:"system"`
}

type ansibleUser struct {
	Append           bool     `yaml:"append,omitempty"`
	Comment          string   `yaml:"comment"`
	CreateHome       bool     `yaml:"create_home,omitempty"`
	Expires          float64  `yaml:"expires,omitempty"`
	Force            bool     `yaml:"force,omitempty"`
	GenerateSSHKey   bool     `yaml:"generate_ssh_key,omitempty"`
	Group            string   `yaml:"group"`
	Groups           []string `yaml:"groups,omitempty"`
	Hidden           bool     `yaml:"hidden,omitempty"`
	Home             string   `yaml:"home"`
	Local            bool     `yaml:"local,omitempty"`
	LoginClass       string   `yaml:"login_class,omitempty"`
	MoveHome         bool     `yaml:"move_home,omitempty"`
	Name             string   `yaml:"name"`
	NonUnique        bool     `yaml:"non_unique,omitempty"`
	Password         string   `yaml:"password,omitempty"`
	PasswordLock     bool     `yaml:"password_lock"`
	Profile          string   `yaml:"profile,omitempty"`
	Remove           bool     `yaml:"remove,omitempty"`
	Role             string   `yaml:"role,omitempty"`
	SEUser           string   `yaml:"seuser,omitempty"`
	Shell            string   `yaml:"shell"`
	Skeleton         string   `yaml:"skeleton,omitempty"`
	SSHKeyBits       int64    `yaml:"ssh_key_bits,omitempty"`
	SSHKeyComment    string   `yaml:"ssh_key_comment,omitempty"`
	SSHKeyFile       string   `yaml:"ssh_key_file,omitempty"`
	SSHKeyPassphrase string   `yaml:"ssh_key_passphrase"`
	SSHKeyType       string   `yaml:"ssh_key_type,omitempty"`
	State            string   `yaml:"state"`
	System           bool     `yaml:"system,omitempty"`
	UID              int      `yaml:"uid"`
	UpdatePassword   string   `yaml:"update_password,omitempty"`
}

type task struct {
	Name          string        `yaml:"name"`
	User          ansibleUser   `yaml:"ansible.builtin.user,omitempty"`
	AuthorizedKey authorizedKey `yaml:"ansible.posix.authorized_key,omitempty"`
	Group         ansibleGroup  `yaml:"ansible.builtin.group,omitempty"`
}

type playbook struct {
	Name              string            `yaml:"name"`
	Hosts             string            `yaml:"hosts"`
	Tasks             []task            `yaml:"tasks"`
	Environment       map[string]string `yaml:"environment"`
	Become            string            `yaml:"become"`
	BecomeMethod      string            `yaml:"become_method,omitempty"`
	IgnoreUnreachable string            `yaml:"ignore_unreachable,omitempty"`
}

type authorizedKey struct {
	Comment       string `yaml:"comment,omitempty"`
	Exclusive     bool   `yaml:"exclusive,omitempty"`
	Follow        bool   `yaml:"follow,omitempty"`
	Key           string `yaml:"key"`
	ManageDir     bool   `yaml:"manage_dir"`
	Path          string `yaml:"path,omitempty,omitempty"`
	State         string `yaml:"state"`
	User          string `yaml:"user"`
	ValidateCerts bool   `yaml:"validate_certs"`
}

func createPlaybook(um *userMap, gm *groupMap, n node) playbook {

	ts := []task{}
	ts = append(ts, groupPlaybook(gm, n)...)
	ts = append(ts, userPlaybook(um, n)...)
	ts = append(ts, sshPlaybook(um, n)...)
	pb := playbook{
		Name:   fmt.Sprintf("%s (%s/%s)", n.Name, n.OS, n.Distro),
		Hosts:  n.Name,
		Tasks:  ts,
		Become: "yes",
		Environment: map[string]string{
			"PATH": "{{ ansible_env.PATH }}:/opt/local/bin:/usr/pkg/bin:/usr/local/bin",
		},
	}
	return pb
}

func sshPlaybook(um *userMap, n node) []task {
	pb := []task{}

	exclude := map[string]bool{}
	for _, u := range n.ExcludeUsers {
		exclude[u] = true
	}

	for _, u := range um.Users {
		if exclude[u.Name] {
			continue
		}

		key := fmt.Sprintf("http://github.com/%s.keys", u.GitHub)

		pb = append(pb, task{
			Name: fmt.Sprintf("ssh key for %s on %s", u.Name, n.Name),
			AuthorizedKey: authorizedKey{
				User:      u.Name,
				Key:       key,
				State:     "present",
				Exclusive: false,
				ManageDir: true,
			},
		})

	}
	return pb
}

func groupPlaybook(gm *groupMap, n node) []task {
	pb := []task{}
	for k, id := range gm.Groups {
		pb = append(pb, task{
			Name: fmt.Sprintf("group %s on %s", k, n.Name),
			Group: ansibleGroup{
				Name:   k,
				ID:     id,
				State:  "present",
				System: false,
			},
		})
	}
	return pb
}

func userPlaybook(um *userMap, n node) []task {
	pb := []task{}

	// On macOS, This will require a volume to be defined in /etc/synthetic.conf
	uroot := "/u"
	shellbin := "/usr/bin"
	password := "*"
	if n.OS == "Darwin" {
		shellbin = "/opt/homebrew/bin"
		password = "*************"
	}
	if n.OS == "FreeBSD" {
		shellbin = "/usr/local/bin"
	}

	if n.OS == "Illumos" {
		shellbin = "/opt/local/bin"
	}

	exclude := map[string]bool{}
	for _, u := range n.ExcludeUsers {
		exclude[u] = true
	}

	for i, u := range um.Users {
		if exclude[u.Name] {
			continue
		}

		if u.Shell == "" {
			u.Shell = *defaultShell
		}

		if u.LoginGroup == "" {
			u.LoginGroup = *defaultGroup
		}

		if n.OS == "Darwin" {
			// "staff" is used for local users on macOS *shrug*
			u.Groups = append(u.Groups, "staff")
		}

		pb = append(pb, task{
			Name: fmt.Sprintf("%s on %s", u.Name, n.Name),
			User: ansibleUser{
				Append:         false,
				CreateHome:     true,
				GenerateSSHKey: true,
				Group:          *defaultGroup,
				Groups:         u.Groups,
				Hidden:         true,
				Home:           fmt.Sprintf("%s/%s", uroot, u.Name),
				PasswordLock:   true,
				Password:       password,
				UID:            *startUID + i,
				Comment:        fmt.Sprintf("%s (%s)", u.Name, u.GitHub),
				Name:           u.Name,
				State:          "present",
				Shell:          fmt.Sprintf("%s/%s", shellbin, u.Shell),
			}})
	}
	return pb
}

// loadUserMap loads a YAML config from a reader
func loadUserMap(r io.Reader) (*userMap, error) {
	bs, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("readall: %w", err)
	}
	klog.Infof("%d bytes read from config", len(bs))

	um := &userMap{}
	err = yaml.Unmarshal(bs, &um)
	if err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	if len(um.Users) == 0 {
		return nil, fmt.Errorf("no entries found after unmarshal")
	}
	klog.Infof("loaded: %+v", um)
	return um, nil
}

// loadGroupMap loads a YAML config from a reader
func loadGroupMap(r io.Reader) (*groupMap, error) {
	bs, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("readall: %w", err)
	}
	klog.Infof("%d bytes read from config", len(bs))

	gm := &groupMap{}
	err = yaml.Unmarshal(bs, &gm)
	if err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	klog.Infof("loaded: %+v", gm)
	if len(gm.Groups) == 0 {
		return nil, fmt.Errorf("no entries found after unmarshal")
	}
	return gm, nil
}

// loadNodeMap loads a YAML config from a reader
func loadNodeMap(r io.Reader) (*nodeMap, error) {
	bs, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("readall: %w", err)
	}
	klog.Infof("%d bytes read from config", len(bs))

	nm := &nodeMap{}
	err = yaml.Unmarshal(bs, &nm)
	if err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	klog.Infof("loaded: %+v", nm)
	if len(nm.Nodes) == 0 {
		return nil, fmt.Errorf("no entries found after unmarshal")
	}
	return nm, nil
}
