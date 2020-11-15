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
		<link href="https://pagecdn.io/lib/chart/2.9.3/Chart.min.css" rel="stylesheet" crossorigin="anonymous"  >
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
                            <li class="list-group-item" data-service="{{.ID}}" data-monitor="{{.MonitorID}}">
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
									<div class="chart-1">
									</div>
									<div class="chart-2">
									</div>
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

		<script src="https://code.jquery.com/jquery-3.5.1.min.js" integrity="sha256-9/aliU8dGd2tb6OSsuzixeV4y/faTqgFtohetphbbj0=" crossorigin="anonymous"></script>
        <script src="https://stackpath.bootstrapcdn.com/bootstrap/4.5.2/js/bootstrap.bundle.min.js" integrity="sha384-LtrjvnR4Twt/qOuYxE721u19sVFLVSA4hf/rRt6PrZTmiPltdZcI7q7PXQBYTKyf" crossorigin="anonymous"></script>
		<script src="https://pagecdn.io/lib/chart/2.9.3/Chart.min.js" crossorigin="anonymous"  ></script>

<script>
$(function() {
	const spinner = $(".spinner-tpl").html();

	$("button.clickable").on("click", function() {
		const when = $(this).data("when");
		const li = $(this).closest(".list-group-item");
		const charts = li.find(".charts:first");
		const chart1 = charts.find(".chart-1");
		const chart2 = charts.find(".chart-2");
		chart1.html(spinner);

		$.get("/api/v1/metrics", { monitor: li.data("monitor"), service: li.data("service"), since: when }, function(result) {
			chart1.html("");
			chart2.html("");
			var canvas1 = document.createElement('canvas');
			var canvas2 = document.createElement('canvas');
			chart1.append(canvas1);
			chart2.append(canvas2);
			var ctx1 = canvas1.getContext('2d');
			var ctx2 = canvas2.getContext('2d');

			var labels1 = [];
			var failedChecks = [];
			var passedChecks = [];
			var labels2 = [];
			var dnsLookups = [];
			var tcpConnections = [];
			var tlsHandshakes = [];
			var serverProcessings = [];
			var contentTransfers = [];

			var day = -1;
			result.forEach(function(item){
				var time = item.bucket;
				var d = new Date(Date.parse(time));
				if (d.getDay() !== day) {
					day = d.getDay();
					time = fmtDate(d.getFullYear(), d.getMonth()+1, d.getDate());
					time += "   ";
					time += fmtTime(d.getHours(), d.getMinutes());
				} else {
					time = fmtTime(d.getHours(), d.getMinutes());
				}

				labels1.push(time);
				failedChecks.push(item.failed_checks);
				passedChecks.push(item.passed_checks);

				labels2.push(time);
				dnsLookups.push(item.details.trace.dns_lookup);
				tcpConnections.push(item.details.trace.tcp_connection);
				tlsHandshakes.push(item.details.trace.tls_handshake);
				serverProcessings.push(item.details.trace.server_processing);
				contentTransfers.push(item.details.trace.content_transfer);
			});

			var datasets1 = [
			{
				label: 'Failed checks',
				backgroundColor: "#dc3545",
				data: failedChecks
			},
			{
				label: 'Passed checks',
				backgroundColor: "#28a745",
				data: passedChecks
			}
			];
			var datasets2 = [
			{
				label: 'DNS lookup',
				backgroundColor: "#31baff",
				data: dnsLookups
			},
			{
				label: 'TCP connection',
				backgroundColor: "#9ceb4f",
				data: tcpConnections
			},
			{
				label: 'TLS handshake',
				backgroundColor: "#ff9326",
				data: tlsHandshakes
			},
			{
				label: 'Server processing',
				backgroundColor: "#9d8cff",
				data: serverProcessings
			},
			{
				label: 'Content transfer',
				backgroundColor: "#18ffe0",
				data: contentTransfers
			},
			];

			new Chart(ctx1, {
				type: 'bar',
				data: {
					labels: labels1,
					datasets: datasets1
				},
				options: {
					title: {
						display: true,
						text: 'HTTP checks'
					},
					tooltips: {
						mode: 'index',
						intersect: false
					},
					responsive: true,
					scales: {
						xAxes: [{
							stacked: true,
						}],
						yAxes: [{
							stacked: true
						}]
					}
				}
			});
			new Chart(ctx2, {
				type: 'bar',
				data: {
					labels: labels2,
					datasets: datasets2
				},
				options: {
					title: {
						display: true,
						text: 'Request tracing [μs] (Avg per time bucket)'
					},
					tooltips: {
						mode: 'index',
						intersect: false
					},
					responsive: true,
					scales: {
						xAxes: [{
							stacked: true,
						}],
						yAxes: [{
							stacked: true
						}]
					}
				}
			});
		}).fail(function() {
			chart1.text("Failed to fetch data");
			chart2.text("");
		});
	});
});

function fmtDate(y, m, d) {
	s = padTime(y);
	s += "-";
	s += padTime(m);
	s += "-";
	s += padTime(d);
	return s;
}

function fmtTime(h, m) {
	s = padTime(h);
	s += ":";
	s += padTime(m);
	return s;
}

function padTime(x) {
	s = "";
	if (x < 10) {
		s += "0";
	}
	s += x;
	return s;
}
</script>
    </body>
</html>
{{end}}
`
