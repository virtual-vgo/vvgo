$(document).ready(function () {
    $('#parts').DataTable({
        paging: false,
        order: [[0, 'asc'], [1, 'asc']],
        columnDefs: [
            {
                targets: [1],
                visible: false,
            },
            { // dont search or order the download links
                targets: [3],
                orderable: false,
                searchable: false,
            },
            {
                className: "text-center",
                targets: "_all"
            }
        ]
    });
});
