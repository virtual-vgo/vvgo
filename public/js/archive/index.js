$(document).ready(function () {
    $('#project_index').DataTable({
        dom: 't',
        paging: false,
        order: [[1, 'asc'], [0, 'asc']],
        columnDefs: [
            {
                targets: [0],
                visible: false,
            }, {
                targets: [1],
                visible: false,
            }, {
                targets: [2],
                className: "text-left",
            }, {
                targets: [3],
                className: "text-left",
            },
        ]
    });
});
