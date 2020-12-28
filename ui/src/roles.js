const axios = require('axios').default;

function GetRoles() {
    const params = new URLSearchParams(window.location.search)
    const paramRoles = params.getAll("roles")
    return axios.post('/roles', paramRoles)
}

export default GetRoles
