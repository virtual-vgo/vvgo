$(document).ready(function () {
    $('#sheets').DataTable({
        paging: false,
        columnDefs: [{"className": "dt-center", "targets": "_all"}],
    });
});
