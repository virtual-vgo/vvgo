function copyPasta(id) {
    let copyText = document.getElementById(id);
    copyText.select();
    document.execCommand("copy");
}
