{{define "body"}}
<div class="button">
<a href="/">首页</a>
</div>

<div>
<h3>用户列表</h3>
<table class="table">
<tr>
    <th class="th1">用户</th>
    <th class="th2">操作</th>
</tr>

{{range $.Users}}
    <tr>
        <td class="name">{{.User}}</td>
        <td>
            <a href="/__gitflow__/useredit?uid={{.Id}}">编辑</a>
            &nbsp;&nbsp;&nbsp;&nbsp;
            <a href="/__gitflow__/userdel?uid={{.Id}}">删除</a>
        </td>
    </tr>
{{end}}
    <tr>
        <td class="name"></td>
        <td>
            <a href="/__gitflow__/useradd">添加用户</a>
        </td>
    </tr>

</table>
</div>

{{end}}
