<% template "admin-header.html" .%>


<span class="pull-right"><a href="/admin/users">Back to list &raquo;</a></span>
<h1>Delete User</h1>

<form method="POST" action="/admin/users/delete/id/<% .user.Id %>">

<p>You are about to delete the user "<% .user.Username %>"</p>

<p>Are you sure you want to do this?</p>

<div class="form-actions">
	<button type="submit" class="btn btn-danger">Delete</button>
	<a href="/admin/users/edit/id/<% .user.Id %>" class="btn">Cancel</a>
</div>
</form>

<% template "admin-footer.html" .%>