<% template "admin-header.html" .%>

<span class="pull-right"><a href="/admin/{{.typeVar}}">Back to list &raquo;</a></span>
<h1>Manage {{.typeVar}}s</h1>
<h3><% if .update %>Edit {{.typeVar}}<% else %>New {{.typeVar}}<% end %></h3>


<div class="row">
  <div class="col-md-6">
  <form role="form" method="POST" action="/admin/{{.typeVar}}/edit<% if .update %>/id/<% .{{.typeVar}}.Id %><% end %>">{{ range $field := .struct }}{{ if $field.IsPk }}
      <% if .{{$.typeVar}}.{{$field.FName}} %>
      <div class="form-group"><label class="control-label" for="{{$field.FName}}">{{$field.FName}}</label>
      <p class="form-control-static"><% .{{$.typeVar}}.{{$field.FName}} %></p>
      </div>
      <% end %>{{else}}{{ if equal $field.DisplayType "text"}}<div class="form-group"><label class="control-label" for="{{$field.FName}}">{{$field.FName}}</label><input 
        type="text" 
        placeholder="{{$field.FName}}" 
        class="form-control" 
        id="{{$field.FName}}" 
        name="{{$field.FName}}" 
        value="<% if .{{$.typeVar}}.{{$field.FName}} %><% .{{$.typeVar}}.{{$field.FName}} %><% end %>"></div>{{end}} 
      {{ if equal $field.DisplayType "checkbox"}}<div class="checkbox"><label><input 
        type="checkbox"         
        id="{{$field.FName}}" 
        name="{{$field.FName}}" 
        value="true" <% if equal .{{$.typeVar}}.{{$field.FName}} "true" %> checked<% end %>>{{$field.FName}}</label></div>{{end}}         
      {{ if equal $field.DisplayType "number"}}<div class="form-group"><label class="control-label" for="{{$field.FName}}">{{$field.FName}}</label><input 
        type="number" 
        placeholder="{{$field.FName}}" 
        class="form-control" 
        id="{{$field.FName}}" 
        name="{{$field.FName}}" 
        value="<% if .{{$.typeVar}}.{{$field.FName}} %><% .{{$.typeVar}}.{{$field.FName}} %><% end %>"></div>{{end}}                
      {{ if equal $field.DisplayType "date"}}<div class="form-group"><label class="control-label" for="{{$field.FName}}">{{$field.FName}}</label><input 
        type="text" 
        placeholder="{{$field.FName}}" 
        class="form-control datepicker"
        data-date-format="mm/dd/yyyy"         
        id="{{$field.FName}}" 
        name="{{$field.FName}}" 
        value="<% if .{{$.typeVar}}.{{$field.FName}} %><% dateformat  .{{$.typeVar}}.{{$field.FName}} "01/02/2006" %><% end %>"></div>{{end}}{{end}}{{ end }}
    <div class="form-group">
      <button type="submit" class="btn btn-primary">Save</button>
      <% if .update %><a href="/admin/{{.typeVar}}/delete/id/<% .{{.typeVar}}.Id %>" type="submit" class="btn btn-danger">Delete</a><% end %>
       <a class="btn btn-default" href="/admin/{{.typeVar}}">Cancel</a>      
    </div>
  
</form>

  </div>
</div>    
<% template "admin-footer.html" .%>