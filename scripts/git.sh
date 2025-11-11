git reset "$(git merge-base master "$(git branch --show-current)")"
git add -A && git commit -m '合并所有提交,初始化第一版代码'
#git pull
#git push --force
# git remote set-url origin https://<你的令牌>@github.com/<你的git用户名>/<要修改的仓库名>.git
# git reset --hard HEAD 回到最后一次提交之前
# git config --global http.proxy 'http://127.0.0.1:7890'
# git config --global https.proxy 'http://127.0.0.1:7890'
