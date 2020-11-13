package deerstatic

// IndexTpl is a root page.
const IndexTpl = `
{{define "index"}}
<!doctype html>
<html lang="en">
    <head>
        <meta charset="utf-8">
        <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">

        <link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.5.2/css/bootstrap.min.css" integrity="sha384-JcKb8q3iqJ61gNV9KGb8thSsNjpSL0n8PARn9HuZOnIxN0hoP+VmmDGMN5t9UJ0Z" crossorigin="anonymous">
        <title>Oh! Deer!</title>
    </head>

    <body>
        <div class="container">
            {{range .Monitors}}
            <div class="row">
                <div class="col">
					<p>
					    <strong>{{.Name}}</strong>
					</p>
                    <ul class="list-group">
                        {{range .Services}}
                            <li class="list-group-item">
                                <strong>{{.Name}}</strong>
								<p class="text-center">
                                {{range .Health}}
                                    {{ if eq .Health 1.0 }}
									<button type="button" class="clickable btn btn-success" data-when="{{.When}}">&nbsp;</button>
                                    {{ else if eq .Health -1.0 }}
									<button type="button" class="btn btn-secondary" data-when="{{.When}}">&nbsp;</button>
                                    {{ else }}
									<button type="button" class="clickable btn btn-danger" data-when="{{.When}}" data-health="{{.Health}}">&nbsp;</button>
                                    {{end}}
                                {{end}}
								</p>
								<div class="charts text-center">
								</div>
                            </li>
                        {{end}}
                    </ul>
                </div>
            </div>
            {{end}}
        </div>

		<div class="spinner-tpl d-none">
			<div class="spinner-border" role="status">
			  <span class="sr-only">Loading...</span>
			</div>
		</div>

		<script
			  src="https://code.jquery.com/jquery-3.5.1.slim.min.js"
			  integrity="sha256-4+XzXVhsDmqanXGHaHvgh1gMQKX40OUvDEBTu8JcmNs="
			  crossorigin="anonymous"></script>
        <script src="https://stackpath.bootstrapcdn.com/bootstrap/4.5.2/js/bootstrap.bundle.min.js" integrity="sha384-LtrjvnR4Twt/qOuYxE721u19sVFLVSA4hf/rRt6PrZTmiPltdZcI7q7PXQBYTKyf" crossorigin="anonymous"></script>

		<script>
$(function() {
	const spinner = $(".spinner-tpl").html();

	$("button.clickable").on("click", function() {
		const charts = $(this).closest(".list-group-item").find(".charts:first");
		console.log(charts);
		charts.html(spinner);
	});
});
		</script>
    </body>
</html>
{{end}}
`
