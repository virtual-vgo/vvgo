import {useEffect, useState} from "react";
import {discordLogin} from "../../auth";
import {RedirectLoginFailure, RedirectLoginSuccess} from "./Login";

export const LoginDiscord = () => {
    const [success, setSuccess] = useState(false);
    const [failed, setFailed] = useState(false);

    const params = new URLSearchParams(window.location.search);
    const code = params.get("code") ?? "";
    const state = params.get("state") ?? "";

    useEffect(() => {
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
