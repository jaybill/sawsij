<%if .info %><div class="alert alert-info"><% .info %></div><% end %>
<%if .success %><div class="alert alert-success"><% .success %></div><% end %>
<%if .errors %>
	<%range $error := .errors%>
	<div class="alert alert-error"><% $error %></div>
	<% end %>
<% end %>
