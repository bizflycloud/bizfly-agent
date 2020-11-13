#!/bin/bash
# (C) BizFly Cloud - 2020
# All rights reserved
# BizFly Cloud Watcher Agent installation script: install and set up the Agent on supported Linux distributions

set -e

function on_error() {
  log_err "\033[31m$ERROR_MESSAGE
It looks like you hit an issue when trying to install the Agent.

Troubleshooting and basic usage information for the Agent are available at:

    https://docs.bizflycloud.vn

If you're still having problems, please send an email to support@bizflycloud.com
with the contents of BAagent-install.log and we'll do our very best to help you
solve your problem.\n\033[0m\n"
}
trap on_error ERR

function usage() {
  this=$1
  cat <<EOF
$this: download go binaries for bizflycloud/bizfly-agent

Usage: $this [-b] bindir [tag]
  -b sets bindir or installation directory, Defaults to ./bin
   [tag] is a tag from
   https://github.com/bizflycloud/bizfly-agent/releases
   If tag is missing, then the latest will be used.

EOF
  exit 2
}

function parse_args() {
  #BINDIR is ./bin unless set be ENV
  # over-ridden by flag below

  BINDIR=${BINDIR:-/usr/bin}
  while getopts "b:dh?x" arg; do
    case "$arg" in
      b) BINDIR="$OPTARG" ;;
      d) log_set_priority 10 ;;
      h | \?) usage "$0" ;;
      x) set -x ;;
    esac
  done
  shift $((OPTIND - 1))
  TAG=$1
}
# this function wraps all the destructive operations
# if a curl|bash cuts off the end of the script due to
# network, either nothing will happen or will syntax error
# out preventing half-done work
function execute() {
  tmpdir=$(mktemp -d)
  log_info "downloading files into ${tmpdir}"
  http_download "${tmpdir}/${TARBALL}" "${TARBALL_URL}"
  http_download "${tmpdir}/${CHECKSUM}" "${CHECKSUM_URL}"
  hash_sha256_verify "${tmpdir}/${TARBALL}" "${tmpdir}/${CHECKSUM}"
  srcdir="${tmpdir}"
  (cd "${tmpdir}" && untar "${TARBALL}")
  test ! -d "${BINDIR}" && install -d "${BINDIR}"
  for binexe in $BINARIES; do
    if [ "$OS" = "windows" ]; then
      binexe="${binexe}.exe"
    fi
    install "${srcdir}/${binexe}" "${BINDIR}/"
    log_info "installed ${BINDIR}/${binexe}"
  done
  rm -rf "${tmpdir}"
}
function get_binaries() {
  case "$PLATFORM" in
    linux/386) BINARIES="bizfly-agent" ;;
    linux/amd64) BINARIES="bizfly-agent" ;;
    linux/arm64) BINARIES="bizfly-agent" ;;
    linux/armv6) BINARIES="bizfly-agent" ;;
    windows/386) BINARIES="bizfly-agent" ;;
    windows/amd64) BINARIES="bizfly-agent" ;;
    windows/arm64) BINARIES="bizfly-agent" ;;
    windows/armv6) BINARIES="bizfly-agent" ;;
    *)
      log_crit "platform $PLATFORM is not supported.  Make sure this script is up-to-date and file request at https://github.com/${PREFIX}/issues/new"
      exit 1
      ;;
  esac
}
function tag_to_version() {
  if [ -z "${TAG}" ]; then
    log_info "checking GitHub for latest tag"
  else
    log_info "checking GitHub for tag '${TAG}'"
  fi
  REALTAG=$(github_release "$OWNER/$REPO" "${TAG}") && true
  if test -z "$REALTAG"; then
    log_crit "unable to find '${TAG}' - use 'latest' or see https://github.com/${PREFIX}/releases for details"
    exit 1
  fi
  # if version starts with 'v', remove it
  TAG="$REALTAG"
  VERSION=${TAG#v}
}
function adjust_format() {
  # change format (tar.gz or zip) based on OS
  case ${OS} in
    windows) FORMAT=zip ;;
  esac
  true
}
function adjust_os() {
  # adjust archive name based on OS
  case ${OS} in
    386) OS=i386 ;;
    amd64) OS=x86_64 ;;
    darwin) OS=Darwin ;;
    linux) OS=Linux ;;
    windows) OS=Windows ;;
  esac
  true
}
function adjust_arch() {
  # adjust archive name based on ARCH
  case ${ARCH} in
    386) ARCH=i386 ;;
    amd64) ARCH=x86_64 ;;
    darwin) ARCH=Darwin ;;
    linux) ARCH=Linux ;;
    windows) ARCH=Windows ;;
  esac
  true
}

function init_config(){
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


  ETCDIR=$(echo "/etc/bizfly-agent")
  CONF="$ETCDIR/bizfly-agent.yaml"
  SYSTEMD_FILE="/lib/systemd/system/bizfly-agent.service"

  GITHUB_ETC_FILE=${RAW_GITHUB_CONTENT}/${CONF}
  GITHUB_SYSTEMD_FILE=${RAW_GITHUB_CONTENT}/scripts/linux/bizfly-agent.service

  BA_API="https://$region.manage.bizflycloud.vn/api/alert"
  pushGW_URL="https://metrics-$region.manage.bizflycloud.vn"
}

function config(){
  # Set the configuration
  if [ -e $CONF ]; then
    printf "\033[34m\n* Keeping old $CONF configuration file\n\033[0m\n"
  else
    if [ ! -d $ETCDIR ]; then
      printf "\033[34m\n* Creating $ETCDIR folder\n\033[0m\n"
      $sudo_cmd mkdir "$ETCDIR"
    fi
    if [ ! -e $CONF ]; then
      printf "\033[34m\n* Getting example configuration file\n\033[0m\n"
      http_download "$CONF" "$GITHUB_ETC_FILE"
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
}

function create_service(){
  # Set the system service
  if [ ! -e $SYSTEMD_FILE ]; then
    http_download "$SYSTEMD_FILE" "$GITHUB_SYSTEMD_FILE"
  fi

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
}

function final(){
  # Metrics are submitted, echo some instructions and exit
  printf "\033[32m

Your Agent is running and functioning properly. It will continue to run in the
background and submit metrics to Datadog.

If you ever want to stop the Agent, run:

    $stop_instructions

And to run it again run:

    $start_instructions

\033[0m"
}

cat /dev/null <<EOF
------------------------------------------------------------------------
portable posix shell functions
------------------------------------------------------------------------
EOF
function is_command() {
  command -v "$1" >/dev/null
}
function echoerr() {
  echo "$@" 1>&2
}
function log_prefix() {
  echo "$0"
}
_logp=6
function log_set_priority() {
  _logp="$1"
}
function log_priority() {
  if test -z "$1"; then
    echo "$_logp"
    return
  fi
  [ "$1" -le "$_logp" ]
}
function log_tag() {
  case $1 in
    0) echo "emerg" ;;
    1) echo "alert" ;;
    2) echo "crit" ;;
    3) echo "err" ;;
    4) echo "warning" ;;
    5) echo "notice" ;;
    6) echo "info" ;;
    *) echo "$1" ;;
  esac
}

function log_info() {
  log_priority 6 || return 0
  echoerr "$(log_prefix)" "$(log_tag 6)" "$@"
}
function log_err() {
  log_priority 3 || return 0
  echoerr "$(log_prefix)" "$(log_tag 3)" "$@"
}
function log_crit() {
  log_priority 2 || return 0
  echoerr "$(log_prefix)" "$(log_tag 2)" "$@"
}
function uname_os() {
  os=$(uname -s | tr '[:upper:]' '[:lower:]')
  case "$os" in
    cygwin_nt*) os="windows" ;;
    mingw*) os="windows" ;;
    msys_nt*) os="windows" ;;
  esac
  echo "$os"
}
function uname_arch() {
  arch=$(uname -m)
  case $arch in
    x86_64) arch="amd64" ;;
    x86) arch="386" ;;
    i686) arch="386" ;;
    i386) arch="386" ;;
    aarch64) arch="arm64" ;;
    armv5*) arch="armv5" ;;
    armv6*) arch="armv6" ;;
    armv7*) arch="armv7" ;;
  esac
  echo ${arch}
}
function uname_os_check() {
  os=$(uname_os)
  case "$os" in
    darwin) return 0 ;;
    dragonfly) return 0 ;;
    freebsd) return 0 ;;
    linux) return 0 ;;
    android) return 0 ;;
    nacl) return 0 ;;
    netbsd) return 0 ;;
    openbsd) return 0 ;;
    plan9) return 0 ;;
    solaris) return 0 ;;
    windows) return 0 ;;
  esac
  log_crit "uname_os_check '$(uname -s)' got converted to '$os' which is not a GOOS value. Please file bug at https://github.com/client9/shlib"
  return 1
}
function uname_arch_check() {
  arch=$(uname_arch)
  case "$arch" in
    386) return 0 ;;
    amd64) return 0 ;;
    arm64) return 0 ;;
    armv5) return 0 ;;
    armv6) return 0 ;;
    armv7) return 0 ;;
    ppc64) return 0 ;;
    ppc64le) return 0 ;;
    mips) return 0 ;;
    mipsle) return 0 ;;
    mips64) return 0 ;;
    mips64le) return 0 ;;
    s390x) return 0 ;;
    amd64p32) return 0 ;;
  esac
  log_crit "uname_arch_check '$(uname -m)' got converted to '$arch' which is not a GOARCH value.  Please file bug report at https://github.com/client9/shlib"
  return 1
}
function untar() {
  tarball=$1
  case "${tarball}" in
    *.tar.gz | *.tgz) tar --no-same-owner -xzf "${tarball}" ;;
    *.tar) tar --no-same-owner -xf "${tarball}" ;;
    *.zip) unzip "${tarball}" ;;
    *)
      log_err "untar unknown archive format for ${tarball}"
      return 1
      ;;
  esac
}
function http_download_curl() {
  local_file=$1
  source_url=$2
  header=$3
  if [ -z "$header" ]; then
    code=$(curl -w '%{http_code}' -sL -o "$local_file" "$source_url")
  else
    code=$(curl -w '%{http_code}' -sL -H "$header" -o "$local_file" "$source_url")
  fi
  if [ "$code" != "200" ]; then
    log_info "http_download_curl received HTTP status $code"
    return 1
  fi
  return 0
}
function http_download_wget() {
  local_file=$1
  source_url=$2
  header=$3
  if [ -z "$header" ]; then
    wget -q -O "$local_file" "$source_url"
  else
    wget -q --header "$header" -O "$local_file" "$source_url"
  fi
}
function http_download() {
  log_info "http_download $2"
  if is_command curl; then
    http_download_curl "$@"
    return
  elif is_command wget; then
    http_download_wget "$@"
    return
  fi
  log_crit "http_download unable to find wget or curl"
  return 1
}
function http_copy() {
  tmp=$(mktemp)
  http_download "${tmp}" "$1" "$2" || return 1
  body=$(cat "$tmp")
  rm -f "${tmp}"
  echo "$body"
}
function github_release() {
  owner_repo=$1
  version=$2
  test -z "$version" && version="latest"
  giturl="https://github.com/${owner_repo}/releases/${version}"
  json=$(http_copy "$giturl" "Accept:application/json")
  test -z "$json" && return 1
  version=$(echo "$json" | tr -s '\n' ' ' | sed 's/.*"tag_name":"//' | sed 's/".*//')
  test -z "$version" && return 1
  echo "$version"
}
function hash_sha256() {
  TARGET=${1:-/dev/stdin}
  if is_command gsha256sum; then
    hash=$(gsha256sum "$TARGET") || return 1
    echo "$hash" | cut -d ' ' -f 1
  elif is_command sha256sum; then
    hash=$(sha256sum "$TARGET") || return 1
    echo "$hash" | cut -d ' ' -f 1
  elif is_command shasum; then
    hash=$(shasum -a 256 "$TARGET" 2>/dev/null) || return 1
    echo "$hash" | cut -d ' ' -f 1
  elif is_command openssl; then
    hash=$(openssl -dst openssl dgst -sha256 "$TARGET") || return 1
    echo "$hash" | cut -d ' ' -f a
  else
    log_crit "hash_sha256 unable to find command to compute sha-256 hash"
    return 1
  fi
}
function hash_sha256_verify() {
  TARGET=$1
  checksums=$2
  log_info "hash_sha256_verify checksum $checksums for $TARGET"
  if [ -z "$checksums" ]; then
    log_err "hash_sha256_verify checksum file not specified in arg2"
    return 1
  fi
  BASENAME=${TARGET##*/}
  want=$(grep "${BASENAME}" "${checksums}" 2>/dev/null | tr '\t' ' ' | cut -d ' ' -f 1)
  if [ -z "$want" ]; then
    log_err "hash_sha256_verify unable to find checksum for '${TARGET}' in '${checksums}'"
    return 1
  fi
  got=$(hash_sha256 "$TARGET")
  if [ "$want" != "$got" ]; then
    log_err "hash_sha256_verify checksum for '$TARGET' did not verify ${want} vs $got"
    return 1
  fi
}
cat /dev/null <<EOF
------------------------------------------------------------------------
End of functions
------------------------------------------------------------------------
EOF

PROJECT_NAME="bizfly-agent"
OWNER=bizflycloud
REPO="bizfly-agent"
BINARY=bizfly-agent
FORMAT=tar.gz
OS=$(uname_os)
ARCH=$(uname_arch)
PREFIX="$OWNER/$REPO"

# use in logging routines
function log_prefix() {
  echo "$PREFIX"
}
PLATFORM="${OS}/${ARCH}"
GITHUB_BRANCH="master"
GITHUB_DOWNLOAD=https://github.com/${OWNER}/${REPO}/releases/download
RAW_GITHUB_CONTENT=https://raw.githubusercontent.com/${PREFIX}/${GITHUB_BRANCH}


uname_os_check "$OS"
uname_arch_check "$ARCH"

parse_args "$@"

get_binaries

tag_to_version

adjust_format

adjust_os

adjust_arch

log_info "found version: ${VERSION} for ${TAG}/${OS}/${ARCH}"

NAME=${PROJECT_NAME}_${OS}_${ARCH}
TARBALL=${NAME}.${FORMAT}
TARBALL_URL=${GITHUB_DOWNLOAD}/${TAG}/${TARBALL}
CHECKSUM=${PROJECT_NAME}_checksums.txt
CHECKSUM_URL=${GITHUB_DOWNLOAD}/${TAG}/${CHECKSUM}


execute


# Root user detection
if [ "$(echo "$UID")" = "0" ]; then
    sudo_cmd=''
else
    sudo_cmd='sudo'
fi


init_config

config

create_service

final
