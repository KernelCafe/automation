# kernel.cafe automation (alpha)

## Adding a node

Adding a node involves 3 steps:

* Create a `barista` user with password-less `sudo` privileges, to be used by [ansible](https://www.ansible.com)
* (Optional) Install `kconfsync`, which will automatically backup system config files to GitHub, omitting keys and passwords.
* Adding the host to DNS

### Automatic

This will run the setup steps automatically on some platforms (Linux, macOS). It's brand new, so it may take some modifications to work properly.

1. `git clone https://github.com/KernelCafe/automation.git /tmp/automation`
2. `cd /tmp/automation/cmd/prepare-node`
3. `sudo ./prep-node.sh`

If it fails, you can run parts of the automation, such as `prep-ansible.sh` and `prep-kconfsync.sh` manually.

### Manual

### 1. Create the barista user

On Linux, run:

```
sudo groupadd -g 2000 barista
sudo useradd -m -g barista -r barista
```

### 2. (Optional) Setup automatic system configuration backups

To track changes to your system configuration, we have a script that syncs changes to directories like `/etc` and `/usr/local/etc` to a GitHub repository. It is recommended that you configure this before introducing further changes.

```
sudo su - barista
cd $HOME
git clone https://github.com/KernelCafe/automation.git
cat /dev/zero | ssh-keygen -q -N ""
git config --global user.email "$(hostname -s)@kernel.cafe"
git config --global user.name "$(hostname -s)"
```

1. Add the contents of $HOME/.ssh/id_rsa.pub` to https://github.com/settings/keys
2. Create a GitHub repository: we typically do so as `KernelCafe/host-$(hostname -s)`

Then run, as the barista user:

```
cd $HOME
repo=host-$(hostname -s)
git clone git@github.com:KernelCafe/$repo.git
cd $repo
cp ../automation/cmd/kconfsync/kconfsync.sh sync.sh
cp ../automation/cmd/kconfsync/gitignore .gitignore
./sync.sh
```

If it works, then install the crontab, as the barista user:

```
tf=$(mktemp)
echo "*/5 * * * * $HOME/${repo}/sync.sh" > $tf
crontab $tf
```

### 3. Give the 'barista' user access to sudo

As an administrative user, run `visudo` and add a line that gives the barista user access to run commands as root:

`barista         ALL = (ALL) NOPASSWD: ALL`

#### 4. Enable 'ansible' to login as the 'barista' user

```
sudo su - barista
mkdir -p $HOME/.ssh
cp $HOME/automation/cmd/prepare-node/barista.pub $HOME/.ssh/authorized_keys
chmod 700 $HOME/.ssh
chmod 400 $HOME/.ssh/*
```

#### 5. Add to node list

To get your node hooked into DNS and ansible, send us PR's to update https://github.com/KernelCafe/welcome/blob/main/nodes/nodes.yaml and  https://github.com/KernelCafe/automation/blob/main/ansible/hosts

Afterwards, we will:

1. Update DNS records at https://www.name.com/account/domain/details/kernel.cafe#dns
3. Run `ansible-playbook -i hosts playbooks/users.yaml`

