git ls-remote --tags origin
git tag -af v`cat VERSION` -m "Bloomsky version `cat VERSION`"
git push origin --tags
