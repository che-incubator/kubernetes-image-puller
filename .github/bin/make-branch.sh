#!/bin/bash -xe
# Release process automation script. 
# Used to create a branch
BASEBRANCH=main
FORCENEWBRANCH=0 # unless forced, don't create a new branch if one already exists. Use with caution!
REPO=git@github.com:che-incubator/kubernetes-image-puller.git

while [[ "$#" -gt 0 ]]; do
  case $1 in
    '-b'|'--branch') BRANCH="$2"; shift 1;;
    '-bf'|'--branchfrom') BASEBRANCH="$2"; shift 1;;
    '--force') FORCENEWBRANCH=1; shift 0;;
  esac
  shift 1
done

usage ()
{
  echo "Usage:   $0 --branch [new branch to create] --branchfrom [source branch]"
  echo "Example: $0 --branch 7.25.x --branchfrom $BASEBRANCH"
  echo 
  echo "Use --force to delete + recreate an existing branch."
  echo
}

if [[ ! ${BRANCH} ]]; then
  usage
  exit 1
fi

# create new branch off ${BASEBRANCH} (recreate only if --force'd)
if [[ "${BASEBRANCH}" != "${BRANCH}" ]]; then
  git fetch
  git branch -D "${BRANCH}" || true
  git checkout "${BASEBRANCH}" || true
  # if branch exists and FORCENEWBRANCH true, delete from remote before creating new branch
  if [[ $(git ls-remote --heads ${REPO} "refs/heads/${BRANCH}" || true) != "" ]] && [[ ${FORCENEWBRANCH} -eq 1 ]]; then
    git push origin ":${BRANCH}"
  fi
  git branch "${BRANCH}"
  git push origin "${BRANCH}"
fi
