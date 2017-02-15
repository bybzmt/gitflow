{{define "body"}}
<div class="confirm_msg">
    {{$.Msg}}
</div>
<div class="confirm_button">
    <a href="{{$.Yes}}">确定</a>

    {{if ne $.No ""}}
        &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;

        <a href="{{$.No}}">取消</a>
    {{end}}
</div>
{{end}}
