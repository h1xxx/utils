date_time=`date +%Y-%m-%d_%H.%M.%S`

col_green='printf \033[01;32m'
col_white='printf \033[01m'
col_reset='printf \033[00m'


echof() {
	case "$1" in
		'white' )   $col_white;;
		'green' )   $col_green;;
	esac
	echo "$2"
	$col_reset
}


line_up() {
	echo -en "\033[1A"
}

# syntax:   test_ssh <ssh machine name>
# stdout:   '1' - sshd running
#		   '0' - sshd not running
# example:  [ $(test_ssh 192.168.0.1)  == '1' ]
test_ssh() {
	ssh -o BatchMode=yes \
		-o ConnectTimeout=3 \
		-o PubkeyAuthentication=no \
		-o PasswordAuthentication=no \
		-o KbdInteractiveAuthentication=no \
		-o ChallengeResponseAuthentication=no \
		$1 2>&1 | grep -c "Permission denied"
}


