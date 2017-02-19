{{define "body"}}
<div class="button">
<a href="/">首页</a>
<a href="/admin/users">用户管理</a>
</div>

<form action="/admin/repoadd_do" method="post">
<div>
<p>
    <label for="name">目录名：</label>
    <input id="name" name="name" type="text" />
</p>
<p>
    <label for="about">说明：</label>
    <input id="about" name="about" type="text" />
</p>
<p>
    <label for="branch_names">分支命名规则：</label><br />
    <textarea id="branch_names" name="branch_names" /></textarea> <br />
    <span>每行1条，可以用正则。空则充许所有</span>
</p>
<p>
    <label for="tag_names">tag规则：</label><br />
    <textarea id="tag_names" name="tag_names" /></textarea> <br />
    <span>每行1条，可以用正则。空则充许所有</span>
</p>
<p>
    <label for="branch_locks">锁定分支列表：</label><br />
    <textarea id="branch_locks" name="branch_locks" /></textarea><br />
    <span>每行1条，可以用正则。空则没有锁定分支</span>
</p>
</div>

<input type="submit" name="act" value="保存" />
&nbsp;&nbsp;&nbsp;&nbsp;
<a href="/">取消</a>

</form>

{{end}}
