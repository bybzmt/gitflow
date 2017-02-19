#!/bin/sh

#会话id
sid="{{$.Sid}}"
hook_url="{{$.HookUrl}}"

# --- Command line
refname="$1"
oldrev="$2"
newrev="$3"


# --- Safety check
if [ -z "$GIT_DIR" ]; then
	echo "Don't run this script from the command line." >&2
	echo " (if you want, you could supply GIT_DIR then run" >&2
	echo "  $0 <ref> <oldrev> <newrev>)" >&2
	exit 1
fi

if [ -z "$refname" -o -z "$oldrev" -o -z "$newrev" ]; then
	echo "usage: $0 <ref> <oldrev> <newrev>" >&2
	exit 1
fi

# --- Check types
# if $newrev is 0000...0000, it's a commit to delete a ref.
zero="0000000000000000000000000000000000000000"
if [ "$newrev" = "$zero" ]; then
	newrev_type=delete
else
	newrev_type=$(git cat-file -t $newrev)
fi

ok=$(curl -s -d "sid=${sid}&refname=${refname}&oldrev=${oldrev}&newrev=${newrev}&newrev_type=${newrev_type}" "${hook_url}")
if [[ $ok != "ok" ]]; then
	echo $ok
	exit 1
fi

# --- Finished
exit 0
