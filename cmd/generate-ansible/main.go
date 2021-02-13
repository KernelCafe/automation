// Sync a YAML map of username:github-user to cloud-init user data

package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v3"
	"k8s.io/klog/v2"
)

var (
	userMapPath     = flag.String("usermap-path", "", "path to usermap")
	groupMapPath    = flag.String("groupmap-path", "", "path to groupmap")
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
}

type userMap struct {
	Users []userConfig `yaml:"users"`
}

type groupMap struct {
	Groups map[string]int64 `yaml:"groups"`
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

	fmt.Println(`---
- name: talk to all hosts just so we can learn about them
  hosts: all
  tasks:
    - name: Classify hosts depending on their OS distribution
      group_by:
        key: os_{{ ansible_facts['system'] }}
`)

	pbs := []playbook{}
	for _, uname := range []string{"Darwin", "Linux"} {
		pbs = append(pbs, createPlaybook(um, gm, uname))
	}

	bs, err := yaml.Marshal(pbs)
	if err != nil {
		klog.Exitf("marshal err: %v", err)
	}
	fmt.Print(string(bs))
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
	Password         string   `yaml:"password"`
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
	Name  string `yaml:"name"`
	Hosts string `yaml:"hosts"`
	Tasks []task `yaml:"tasks"`
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

func createPlaybook(um *userMap, gm *groupMap, uname string) playbook {
	ts := []task{}
	ts = append(ts, groupPlaybook(gm)...)
	ts = append(ts, userPlaybook(um, uname)...)
	ts = append(ts, sshPlaybook(um)...)
	return playbook{Name: fmt.Sprintf("%s users", uname), Hosts: fmt.Sprintf("os_%s", uname), Tasks: ts}
}

func sshPlaybook(um *userMap) []task {
	pb := []task{}

	for _, u := range um.Users {
		key := fmt.Sprintf("http://github.com/%s.keys", u.GitHub)

		pb = append(pb, task{
			Name: fmt.Sprintf("ssh key for %s", u.Name),
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

func groupPlaybook(gm *groupMap) []task {
	pb := []task{}
	for k, id := range gm.Groups {
		pb = append(pb, task{
			Name: fmt.Sprintf("group for %s", k),
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

func userPlaybook(um *userMap, uname string) []task {
	pb := []task{}

	uroot := "/u"
	shellbin := "/usr/bin"
	if uname == "Darwin" {
		// This will require a volume to be defined in /etc/synthetic.conf
		uroot = "/u"
		shellbin = "/opt/homebrew/bin"
	}

	for i, u := range um.Users {
		if u.Shell == "" {
			u.Shell = *defaultShell
		}

		if u.LoginGroup == "" {
			u.LoginGroup = *defaultGroup
		}

		pb = append(pb, task{
			Name: u.Name,
			User: ansibleUser{
				Append:         false,
				Comment:        u.GitHub,
				CreateHome:     true,
				GenerateSSHKey: true,
				Group:          *defaultGroup,
				Groups:         u.Groups,
				Hidden:         true,
				Home:           fmt.Sprintf("%s/%s", uroot, u.Name),
				PasswordLock:   true,
				Password:       "*",
				UID:            *startUID + i,
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
