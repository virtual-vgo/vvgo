$(document).ready(function () {
    $('#parts').DataTable({
        scrollY: '50vh',
        scrollCollapse: true,
        paging: false,
        "columnDefs": [
            {
                targets: [2],
                orderable: false
            },
            {
                className: "dt-center",
                targets: "_all"
            }
        ]
    });
});
