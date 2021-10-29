import {discordLogin} from "../../auth";
import {RedirectLoginFailure, RedirectLoginSuccess} from "./Login";
import _ = require("lodash");
import React = require("react");

export const LoginDiscord = () => {
    const [success, setSuccess] = React.useState(false);
    const [failed, setFailed] = React.useState(false);

    const params = new URLSearchParams(window.location.search);
    const code = _.defaultTo(params.get("code"), "");
    const state = _.defaultTo(params.get("state"), "");

    React.useEffect(() => {
        discordLogin(code, state)
            .then(me => {
                setSuccess(true);
                console.log("login successful", me);
            })
            .catch(err => {
                setFailed(true);
                console.log("login failed", err);
            });
    });

    switch (true) {
        case success:
            return <RedirectLoginSuccess/>;
        case failed:
            return <RedirectLoginFailure/>;
        default:
            return <div>Loading...</div>;
    }
};