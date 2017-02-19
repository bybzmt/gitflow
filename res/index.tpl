{{define "body"}}
<div class="button">
<a href="/admin/users">用户管理</a>
<a href="/admin/useradd">用户注册</a>
</div>

<div>
<h3>项目列表</h3>
<table class="table">
<tr>
    <th class="th1">名称</th>
    <th class="th2">描术</th>
    <th class="th2">地址</th>
    <th class="th2">操作</th>
</tr>

{{range $.Repos}}
    <tr>
        <td class="name">{{.Name}}</td>
        <td>{{.Message}}</td>
        <td>{{$.ReposBase}}{{.Name}}</td>
        <td>
            <a href="/admin/repoedit?rid={{.Id}}">编辑</a>
            &nbsp;&nbsp;&nbsp;&nbsp;
            <a href="/admin/repodel?rid={{.Id}}">删除</a>
        </td>
    </tr>
{{end}}

    <tr>
        <td class="name"></td>
        <td></td>
        <td></td>
        <td>
            <a href="/admin/repoadd">添加</a>
        </td>
    </tr>

</table>
</div>

{{end}}
