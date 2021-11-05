import {fetchApi, OAuthRedirect, Session} from "../datasets";
import _ = require("lodash");

const SessionItemKey = "session";

export const logout = () => {
    return fetchApi("/auth/logout", {method: "GET"})
        .then(() => setSession({} as Session));
};

export const updateLogin = () => {
    fetchApi("/me", {method: "GET"}).then(resp => {
        const me = _.defaultTo(resp.Identity, {} as Session);
        if (me.Key != getSession().Key) setSession(me);
    });
};

const setSession = (session: Session) => {
    const sessionJSON = JSON.stringify(session);
    localStorage.clear();
    localStorage.setItem(SessionItemKey, sessionJSON);
};

export const getSession = (): Session => {
    const sessionJSON = _.defaultTo(localStorage.getItem(SessionItemKey), "");
    const session = _.isEmpty(sessionJSON) ? {} : JSON.parse(sessionJSON);
    const params = new URLSearchParams(window.location.search);
    if (params.has("token")) session.Key = params.get("token");
    if (params.has("roles")) session.Roles = _.defaultTo(params.get("roles"), "").split(",");
    return session;
};

export const passwordLogin = async (user: string, pass: string): Promise<Session> => {
    const params = new URLSearchParams({user: user, pass: pass});
    return fetchApi("/auth/password?" + params.toString(), {method: "POST"})
        .then(resp => {
            setSession(resp.Identity);
            return resp.Identity;
        });
};

export const oauthRedirect = async (): Promise<OAuthRedirect> => {
    return fetchApi("/auth/oauth_redirect", {method: "GET"})
        .then(resp => {
            const itemKey = "oauth_redirect_secret:" + resp.OAuthRedirect.State;
            localStorage.setItem(itemKey, resp.OAuthRedirect.Secret);
            return resp.OAuthRedirect;
        });
};

export const discordLogin = async (code: string, state: string): Promise<Session> => {
    const itemKey = "oauth_redirect_secret:" + state;
    const secret = localStorage.getItem(itemKey);
    localStorage.removeItem(itemKey);
    const data = {code, state, secret};
    return fetchApi("/auth/discord", {method: "POST", body: JSON.stringify(data)})
        .then(resp => {
            setSession(resp.Identity);
            return resp.Identity;
        });
};
