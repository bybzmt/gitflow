{{define "body"}}
<div class="button">
<a href="/__gitflow__/users">用户管理</a>
</div>

<div>
<h3>项目列表</h3>
<table class="table">
<tr>
    <th class="th1">名称</th>
    <th class="th2">描术</th>
    <th class="th2">操作</th>
</tr>

{{range $.Repos}}
    <tr>
        <td class="name">{{.Name}}</td>
        <td>{{.Message}}</td>
        <td>
            <a href="/__gitflow__/repoedit?rid={{.Id}}">编辑</a>
            &nbsp;&nbsp;&nbsp;&nbsp;
            <a href="/__gitflow__/repodel?rid={{.Id}}">删除</a>
        </td>
    </tr>
{{end}}

    <tr>
        <td class="name"></td>
        <td></td>
        <td>
            <a href="/__gitflow__/repoadd">添加</a>
        </td>
    </tr>

</table>
</div>

{{end}}
