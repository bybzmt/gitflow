{{define "body"}}
<div class="button">
<a href="/">首页</a>
<a href="/admin/users">用户管理</a>
</div>

<form action="/admin/repoedit_do" method="post">
<input id="rid" name="rid" type="hidden" value="{{$.Repo.Id}}" />

<div>
<p><label for="name">目录名：</label><input id="name" name="name" type="text" value="{{$.Repo.Name}}" /></p>
<p><label for="about">说明：</label><input id="about" name="about" type="text" value="{{$.Repo.Message}}" /></p>
</div>

<p>
    <label for="branch_names">分支命名规则：</label><br />
    <textarea id="branch_names" name="branch_names" />{{$.BranchNames}}</textarea> <br />
    <span>每行1条，可以用正则。空则充许所有</span>
</p>
<p>
    <label for="tag_names">tag规则：</label><br />
    <textarea id="tag_names" name="tag_names" />{{$.TagNames}}</textarea> <br />
    <span>每行1条，可以用正则。空则充许所有</span>
</p>
<p>
    <label for="branch_locks">锁定分支列表：</label><br />
    <textarea id="branch_locks" name="branch_locks" />{{$.BranchLocks}}</textarea><br />
    <span>每行1条，可以用正则。空则没有锁定分支</span>
</p>

<input type="submit" name="act" value="保存" />
&nbsp;&nbsp;&nbsp;&nbsp;
<a href="/">取消</a>

</form>

{{end}}
