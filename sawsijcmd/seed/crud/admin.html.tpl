<% template "admin-header.html" .%>

<span class="pull-right"><a class="btn btn-primary" href="/admin/{{.typeVar}}/edit" title="Add new {{.typeVar}}"><i class="icon-plus-sign icon-white"></i> Add New</a></span>
<h1>Manage {{.typeVar}}s</h1>


<% if .{{.typeVar}}s %>
<table class="table table-striped table-bordered table-condensed">
  <thead>
    <tr>
{{ range $field := .struct }}       <th>{{$field.FName}}</th>       
{{ end }}    </tr>
  </thead>
  <tbody>
    <%range $index,${{.typeVar}} := .{{.typeVar}}s %>
    <tr>
    {{ range $i, $field := .struct }}<td>
     
        <a href="/admin/{{ $.typeVar }}/edit/id/<% ${{ $.typeVar}}.Id %>">
            <% ${{$.typeVar}}.{{$field.FName}} %>
        </a>
      </td>       
    {{ end }}
    </tr>      
    <%end%>
  </tbody>
</table>
<% else %>
<div class="alert alert-info">
              <button type="button" class="close" data-dismiss="alert">×</button>
              <strong>No {{.typeVar}}s found.</strong> If you'd like, you can <a href="/admin/{{.typeVar}}/edit">create one</a>.
            </div>
<% end %>
<% template "admin-footer.html" .%>