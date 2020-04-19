$(document).ready(function () {
    $('#parts').DataTable({
        scrollY: "40vh",
        scrollCollapse: true,
        paging: false,
        order: [1, 'asc'],
        columnDefs: [
            {
                targets: [2],
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
