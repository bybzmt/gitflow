{{define "body"}}
<div class="button">
<a href="/">首页</a>
<a href="/admin/users">用户管理</a>
</div>

<form action="/admin/useradd_do" method="post">
<div>
<p><label for="user">用户名：</label><input id="user" name="user" type="text" /></p>
<p><label for="pass">密码：</label><input id="pass" name="pass" type="password" /></p>
</div>

<input type="submit" name="act" value="保存" />
&nbsp;&nbsp;&nbsp;&nbsp;
<a href="/">取消</a>

</form>

{{end}}
