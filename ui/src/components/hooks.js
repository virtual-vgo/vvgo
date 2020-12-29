import {useEffect, useState} from 'react';

const axios = require('axios').default;

function useLoginRoles() {
    const params = new URLSearchParams(window.location.search)
    const paramRoles = params.getAll("roles")

    const [roles, setRoles] = useState([]);
    useEffect(() => {
        axios.post('/roles', paramRoles)
            .then(response => setRoles(response.data))
            .catch(error => console.log(error))
    });
    return roles
}

export {useLoginRoles}
