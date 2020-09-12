$(document).ready(function () {
    $('#parts').DataTable({
        paging: false,
        order: [[0, 'asc']],
        columnDefs: [
            {
                targets: [0],
                visible: false,
            },
            { // dont search or order the download links
                targets: [2],
                orderable: false,
                searchable: false,
            },
            {
                className: "text-left",
                targets: "_all"
            }
        ]
    });
});
