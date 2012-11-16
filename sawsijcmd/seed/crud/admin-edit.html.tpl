<% template "admin-header.html" .%>

<span class="pull-right"><a href="/admin/{{.typeVar}}">Back to list &raquo;</a></span>
<h1>Manage Organizations</h1>

<form class="form-horizontal" method="POST" action="/admin/{{.typeVar}}/edit<% if .update %>/id/<% .{{.typeVar}}.Id %><% end %>">
  <fieldset>
   <legend><% if .update %>Edit {{.typeVar}}<% else %>New {{.typeVar}}<% end %></legend>
    {{ range $field := .struct }}
    <div class="control-group">
      <label class="control-label" for="{{$field.FName}}">{{$field.FName}}</label>
      <div class="controls">
       <input type="text" placeholder="{{$field.FName}}" class="span5" maxlength="64" id="{{$field.FName}}" name="{{$field.FName}}" value="<% if .{{$.typeVar}}.{{$field.FName}} %><% .{{$.typeVar}}.{{$field.FName}} %><% end %>">      
      </div>
    </div>
    {{ end }}
    <div class="form-actions">
      <button type="submit" class="btn btn-primary">Save</button> 
      <% if .update %><a href="/admin/{{.typeVar}}/delete/id/<% .{{.typeVar}}.Id %>" type="submit" class="btn btn-danger">Delete</a><% end %>      
    </div>

  </fieldset>
</form>

<% template "admin-footer.html" .%>