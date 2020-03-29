$(document).ready(function () {
    $('#example').DataTable({
        paging: false,
        columnDefs: [{"className": "dt-center", "targets": "_all"}],
    });
});
