{{define "body"}}
<div class="button">
<a href="/">首页</a>
<a href="/__gitflow__/users">用户管理</a>
</div>

<form action="/__gitflow__/repoadd_do" method="post">
<div>
<p><label for="name">目录名：</label><input id="name" name="name" type="text" /></p>
<p><label for="about">说明：</label><input id="about" name="about" type="text" /></p>
</div>

<input type="submit" name="act" value="保存" />
&nbsp;&nbsp;&nbsp;&nbsp;
<a href="/">取消</a>

</form>

{{end}}
