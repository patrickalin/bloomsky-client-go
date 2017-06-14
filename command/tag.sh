git ls-remote --tags origin
git tag -af v`cat VERSION` -m "my version `cat VERSION`"
git push origin --tags
