$(document).ready(function () {
    $('#parts').DataTable({
        scrollY: "40vh",
        scrollCollapse: true,
        paging: false,
        order: [[1, 'asc'], [2, 'asc']],
        columnDefs: [
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
