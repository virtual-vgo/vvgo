$(document).ready(function () {
    $('#parts').DataTable({
        scrollY: "40vh",
        scrollCollapse: true,
        paging: false,
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
