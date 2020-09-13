$(document).ready(function () {
    $('#parts').DataTable({
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
