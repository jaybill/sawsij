<% template "admin-header.html" .%>


<span class="pull-right"><a href="/admin/{{.typeVar}}">Back to list &raquo;</a></span>
<h1>Delete {{.typeVar}}</h1>

<form method="POST" action="/admin/{{.typeVar}}/delete/id/<% .{{.typeVar}}.Id %>">

<p>You are about to delete this {{.typeVar}}</p>

<p>Are you sure you want to do this?</p>

<div class="form-actions">
	<button type="submit" class="btn btn-danger">Delete</button>
	<a href="/admin/{{.typeVar}}/edit/id/<% .{{.typeVar}}.Id %>" class="btn">Cancel</a>
</div>
</form>

<% template "admin-footer.html" .%>