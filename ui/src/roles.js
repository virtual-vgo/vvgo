const axios = require('axios').default;

function GetRoles() {
    let roles = []
    const params = new URLSearchParams(window.location.search)
    const paramRoles = params.get("roles")
    axios.get('/roles', {params: {roles: paramRoles}})
        .then(response => roles = response.data)
    return roles
}

export default GetRoles
