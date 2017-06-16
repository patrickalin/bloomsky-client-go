git ls-remote --tags origin
git tag -af `cat VERSION` -m "`cat VERSION`"
git push origin --tags
