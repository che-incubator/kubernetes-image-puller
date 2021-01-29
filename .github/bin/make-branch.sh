#!/bin/bash
# Release process automation script. 
# Used to create a branch
BASEBRANCH=main
FORCENEWBRANCH=0 # unless forced, don't create a new branch if one already exists. Use with caution!

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
  git checkout "${BASEBRANCH}" || true
  git branch --set-upstream-to="origin/${BRANCH}" "${BRANCH}" -q || { 
    if [[ ${FORCENEWBRANCH} -eq 0 ]]; then 
      echo "[INFO] Branch ${BRANCH} already exists: nothing to do!"
    else 
      echo "[INFO] Branch ${BRANCH} already exists: deleting and recreating branch"
      git push origin ":${BRANCH}"
      git branch "${BRANCH}"
    fi
  }
  git pull origin "${BRANCH}" || true
  git push origin "${BRANCH}"
fi
