package deerstatic

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
                    <h4>{{.Name}}</h4>
                    <ul class="list-group">
                        {{range .Services}}
                            <li class="list-group-item">
                                <strong>{{.Name}}</strong><br>
                                {{range .Health}}
                                    {{ if eq .Health 1.0 }}
                                    <span class="badge badge-success" title="{{.When}} [Healthy]">&nbsp;</span>
                                    {{ else if eq .Health -1.0 }}
                                    <span class="badge badge-secondary" title="{{.When}} [No data]">&nbsp;</span>
                                    {{ else }}
                                    <span class="badge badge-danger" data-health="{{.Health}}" title="{{.When}} [{{.Health}}]">&nbsp;</span>
                                    {{end}}
                                {{end}}
                            </li>
                        {{end}}
                    </ul>
                </div>
            </div>
            {{end}}
        </div>

        <script src="https://stackpath.bootstrapcdn.com/bootstrap/4.5.2/js/bootstrap.bundle.min.js" integrity="sha384-LtrjvnR4Twt/qOuYxE721u19sVFLVSA4hf/rRt6PrZTmiPltdZcI7q7PXQBYTKyf" crossorigin="anonymous"></script>
    </body>
</html>
{{end}}
`
