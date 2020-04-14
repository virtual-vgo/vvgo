$(document).ready(function () {
    $('#parts').DataTable({
        scrollY: '50vh',
        scrollCollapse: true,
        paging: false,
        "columnDefs": [
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
