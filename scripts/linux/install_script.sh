#!/bin/bash
# (C) BizFly Cloud - 2020
# All rights reserved
# BizFly Alert Agent installation script: install and set up the Agent on supported Linux distributions

set -e
install_script_version=1.0.0
logfile="BAagent-install.log"

ETCDIR=$(echo "/etc/bizfly-agent")
CONF="$ETCDIR/bizfly-agent.yaml"
SERVICE="/lib/systemd/system/bizfly-agent.service"

# Set up a named pipe for logging
npipe=/tmp/$$.tmp
mknod $npipe p

# Log all output to a log for error checking
tee <$npipe $logfile &
exec 1>&-
exec 1>$npipe 2>&1
trap 'rm -f $npipe' EXIT


function on_error() {
    printf "\033[31m$ERROR_MESSAGE
It looks like you hit an issue when trying to install the Agent.

Troubleshooting and basic usage information for the Agent are available at:

    https://docs.bizflycloud.vn

If you're still having problems, please send an email to support@bizflycloud.com
with the contents of BAagent-install.log and we'll do our very best to help you
solve your problem.\n\033[0m\n"
}
trap on_error ERR

secretkey=
if [ -n "$BA_SECRET_KEY" ]; then
    secretkey=$BA_SECRET_KEY
fi

secretid=
if [ -n "$BA_SECRET_ID" ]; then
    secretid=$BA_SECRET_ID
fi

projectid=
if [ -n "$BA_PROJECT_ID" ]; then
    projectid=$BA_PROJECT_ID
fi

no_start=
if [ -n "$BA_INSTALL_ONLY" ]; then
    no_start=true
fi

region=
if [ -n "$REGION" ]; then
    region=$REGION
fi


BA_API="https://$region.manage.bizflycloud.vn/api/alert"
pushGW_URL="https://metrics-$region.manage.bizflycloud.vn"


agent_version=${AGENT_VERSION:-v1.0.1}
install_dir=${INSTALL_DIR:-"/opt/bizfly-agent"}


# OS/Distro Detection
# Try lsb_release, fallback with /etc/issue then uname command
KNOWN_DISTRIBUTION="(Debian|Ubuntu|RedHat|CentOS|openSUSE|Amazon|Arista|SUSE)"
DISTRIBUTION=$(lsb_release -d 2>/dev/null | grep -Eo $KNOWN_DISTRIBUTION  || grep -Eo $KNOWN_DISTRIBUTION /etc/issue 2>/dev/null || grep -Eo $KNOWN_DISTRIBUTION /etc/Eos-release 2>/dev/null || grep -m1 -Eo $KNOWN_DISTRIBUTION /etc/os-release 2>/dev/null || uname -s)

if [ "$DISTRIBUTION" = "Darwin" ]; then
    printf "\033[31mThis script does not support installing on the Mac."
    exit 1;

elif [ -f /etc/debian_version ] || [ "$DISTRIBUTION" == "Debian" ] || [ "$DISTRIBUTION" == "Ubuntu" ]; then
    OS="Debian"
elif [ -f /etc/redhat-release ] || [ "$DISTRIBUTION" == "RedHat" ] || [ "$DISTRIBUTION" == "CentOS" ] || [ "$DISTRIBUTION" == "Amazon" ]; then
    OS="RedHat"
# Some newer distros like Amazon may not have a redhat-release file
elif [ -f /etc/system-release ] || [ "$DISTRIBUTION" == "Amazon" ]; then
    OS="RedHat"
# Arista is based off of Fedora14/18 but do not have /etc/redhat-release
elif [ -f /etc/Eos-release ] || [ "$DISTRIBUTION" == "Arista" ]; then
    OS="RedHat"
# openSUSE and SUSE use /etc/SuSE-release or /etc/os-release
elif [ -f /etc/SuSE-release ] || [ "$DISTRIBUTION" == "SUSE" ] || [ "$DISTRIBUTION" == "openSUSE" ]; then
    OS="SUSE"
fi

# Root user detection
if [ "$(echo "$UID")" = "0" ]; then
    sudo_cmd=''
else
    sudo_cmd='sudo'
fi

# Install the necessary package sources
if [ "$OS" != "SUSE" ] && [ "$OS" != "Debian" ] && [ "$OS" != "RedHat" ]; then
    printf "\033[31mYour OS or distribution are not supported by this install script.
Please follow the instructions on the Agent setup page:

    https://docs.bizflycloud.vn"
    exit;
fi

# Install the necessary package sources
printf "Installing necessary package tools: wget, tar"
if [ "$OS" = "RedHat" ]; then
  $sudo_cmd yum install -y wget tar

elif [ "$OS" = "Debian" ]; then
  $sudo_cmd apt install -y wget tar

elif [ "$OS" = "SUSE" ]; then
  $sudo_cmd rpm install -y wget tar

fi


UNAME_M=$(uname -m)
if [ "$UNAME_M"  == "i686" ] || [ "$UNAME_M"  == "i386" ] || [ "$UNAME_M"  == "x86" ]; then
    ARCHI="i386"
elif [ "$UNAME_M"  == "aarch64" ]; then
    ARCHI="arm64"
else
    ARCHI="x86_64"
fi

FILE_URL=$(echo "https://github.com/bizflycloud/bizfly-agent/releases/download/$agent_version/bizfly-agent_Linux_$ARCHI.tar.gz")
DOWNLOAD_TOOL=`$sudo_cmd which wget`
EXTRACT_TOOL=`$sudo_cmd which tar`

printf "\033[34m\n* Downloading agent package\n\033[0m\n"
$sudo_cmd $DOWNLOAD_TOOL "$FILE_URL" -O "bizfly-agent_Linux_$ARCHI.tar.gz"

printf "Extracting agent package to $install_dir"
if [ ! -d $install_dir ]; then
  $sudo_cmd mkdir "$install_dir"
fi

$sudo_cmd $EXTRACT_TOOL -xvf "bizfly-agent_Linux_$ARCHI.tar.gz" -C "$install_dir"


# Add permission
$sudo_cmd chown -R root:root "$install_dir"
$sudo_cmd chmod 640 "$install_dir"
$sudo_cmd chmod 750 "$install_dir/bizfly-agent"


# Set the configuration
if [ -e $CONF ]; then
  printf "\033[34m\n* Keeping old /etc/bizfly-agent/bizfly-agent.yaml configuration file\n\033[0m\n"
else
  if [ ! -d $ETCDIR ]; then
    printf "\033[34m\n* Creating $ETCDIR folder\n\033[0m\n"
    $sudo_cmd mkdir "$ETCDIR"
  fi
  if [ ! -e $CONF ]; then
    printf "\033[34m\n* Getting example configuration file\n\033[0m\n"
    $sudo_cmd "$DOWNLOAD_TOOL" "https://raw.githubusercontent.com/bizflycloud/bizfly-agent/master/bizfly-agent.yaml" -O "$CONF"
  fi
  if [ "$secretkey" ]; then
    printf "\033[34m\n* Adding your Secret Key to the Agent configuration: $CONF\n\033[0m\n"
    $sudo_cmd sh -c "sed -i 's/secret:.*/secret: $secretkey/' $CONF"
  else
    printf "\033[31mThe Agent won't start automatically at the end of the script because the Secret Key is missing, please add one in bizfly-agent.yaml and start the agent manually.\n\033[0m\n"
    no_start=true
  fi
  if [ -n "$secretid" ]; then
    printf "\033[34m\n* Setting Secret ID in the Agent configuration: $CONF\n\033[0m\n"
    $sudo_cmd sh -c "sed -i 's/secretID:.*/secretID: $secretid/' $CONF"
  else
    printf "\033[31mThe Agent won't start automatically at the end of the script because the Secret ID is missing, please add one in bizfly-agent.yaml and start the agent manually.\n\033[0m\n"
    no_start=true
  fi
  if [ -n "$projectid" ]; then
    printf "\033[34m\n* Setting Project ID in the Agent configuration: $CONF\n\033[0m\n"
    $sudo_cmd sh -c "sed -i 's/project:.*/project: $projectid/' $CONF"
  else
    printf "\033[31mThe Agent won't start automatically at the end of the script because the Project ID is missing, please add one in bizfly-agent.yaml and start the agent manually.\n\033[0m\n"
    no_start=true
  fi
  if [ -n "$region" ]; then
    printf "\033[34m\n* Setting Default Endpoint BizFly Alert API in the Agent configuration: $CONF\n\033[0m\n"
    $sudo_cmd sh -c "sed -i 's|defaultEndpoint:.*|defaultEndpoint: $BA_API|' $CONF"
    $sudo_cmd sh -c "sed -i 's|url:.*|url: $pushGW_URL|' $CONF"
  else
    printf "\033[31mThe Agent won't start automatically at the end of the script because the Region is missing, please add one in bizfly-agent.yaml and start the agent manually.\n\033[0m\n"
    no_start=true
  fi

  $sudo_cmd chown root:root $CONF
  $sudo_cmd chmod 640 $CONF
fi

# Set the system service
if [ ! -e $SERVICE ]; then
  $sudo_cmd "$DOWNLOAD_TOOL" "https://raw.githubusercontent.com/bizflycloud/bizfly-agent/master/scripts/linux/bizfly-agent.service" -O "$SERVICE"
fi

# Creating or overriding the install information
install_info_content="---
install_method:
  tool: install_script
  tool_version: install_script
  installer_version: install_script-$install_script_version
"
$sudo_cmd sh -c "echo '$install_info_content' > $ETCDIR/install_info"

# On SUSE 11, sudo service bizfly-agent start fails (because /sbin is not in a base user's path)
# However, sudo /sbin/service bizfly-agent does work.
# Use which (from root user) to find the absolute path to service
service_cmd="service"
if [ "$SUSE11" == "yes" ]; then
  service_cmd=`$sudo_cmd which service`
fi

# Use /usr/sbin/service by default.
# Some distros usually include compatibility scripts with Upstart or Systemd. Check with: `command -v service | xargs grep -E "(upstart|systemd)"`
restart_cmd="$sudo_cmd $service_cmd bizfly-agent restart"
stop_instructions="$sudo_cmd $service_cmd bizfly-agent stop"
start_instructions="$sudo_cmd $service_cmd bizfly-agent start"

if command -v systemctl 2>&1; then
  # Use systemd if systemctl binary exists
  restart_cmd="$sudo_cmd systemctl restart bizfly-agent.service"
  stop_instructions="$sudo_cmd systemctl stop bizfly-agent"
  start_instructions="$sudo_cmd systemctl start bizfly-agent"
elif /sbin/init --version 2>&1 | grep -q upstart; then
  # Try to detect Upstart, this works most of the times but still a best effort
  restart_cmd="$sudo_cmd stop bizfly-agent || true ; sleep 2s ; $sudo_cmd start bizfly-agent"
  stop_instructions="$sudo_cmd stop bizfly-agent"
  start_instructions="$sudo_cmd start bizfly-agent"
fi

if [ $no_start ]; then
    printf "\033[34m
* BA_INSTALL_ONLY environment variable set: the newly installed version of the agent
will not be started. You will have to do it manually using the following
command:

    $start_instructions

\033[0m\n"
    exit
fi

printf "\033[34m* Starting the Agent...\n\033[0m\n"
eval "$restart_cmd"


# Metrics are submitted, echo some instructions and exit
printf "\033[32m

Your Agent is running and functioning properly. It will continue to run in the
background and submit metrics to Datadog.

If you ever want to stop the Agent, run:

    $stop_instructions

And to run it again run:

    $start_instructions

\033[0m"
