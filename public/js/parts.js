$(document).ready(function () {
    $('a[data-toggle="pill"]').on('shown.bs.tab', function (e) {
        $.fn.dataTable.tables({visible: true, api: true}).columns.adjust();
    });
    $('table.table').DataTable({
        dom: 'ft',
        paging: false,
        order: [[0, 'asc']],
        columnDefs: [
            {
                targets: [0],
                visible: false,
            },
            {
                targets: [1],
                className: "text-left",
            },
            { // dont search or order the download links
                targets: [2],
                orderable: false,
                searchable: false,
                className: "text-left",
            },
        ]
    });
});
