{{define "body"}}
<div class="button">
<a href="/">首页</a>
<a href="/__gitflow__/users">用户管理</a>
</div>

<form action="/__gitflow__/useradd_do" method="post">
<div>
<p><label for="user">用户名：</label><input id="user" name="user" type="text" /></p>
<p><label for="pass">密码：</label><input id="pass" name="pass" type="password" /></p>
<p><label for="isadmin">是否管理员：</label><input id="isadmin" name="isadmin" type="checkbox" value="1" /></p>
</div>

<div>
<h3>权限管理</h3>
<table class="table">
<tr>
    <th class="th1">项目</th>
    {{range $.Rules}}
    <th>{{.About}}</th>
    {{end}}
</tr>

{{range $.Repos}}
    {{$repo:=.}}
    <tr>
        <td class="name">{{.Name}}</td>
        {{range $.Rules}}
        <td><input type="checkbox" name="perms[]" value="{{$repo.Id}}:{{.Id}}" /></td>
        {{end}}
    </tr>
{{end}}

</table>
</div>

<input type="submit" name="act" value="保存" />
&nbsp;&nbsp;&nbsp;&nbsp;
<a href="/__gitflow__/users">取消</a>

</form>

{{end}}
