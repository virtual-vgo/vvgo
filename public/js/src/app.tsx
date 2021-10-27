import React = require("react");
import ReactDOM = require("react-dom");
import {BrowserRouter, Redirect, Route, Switch} from "react-router-dom";
import {getSession} from "./auth";
import {About} from "./components/About";
import {Contact} from "./components/Contact";
import {Home} from "./components/Home";
import {Login, LoginDiscord, LoginFailure, LoginSuccess} from "./components/Login";
import {Sessions} from "./components/Sessions";
import {sessionIsAnonymous, UserRoles} from "./datasets";
import {MemberDashboard} from "./mixtape/MemberDashboard";
import {NewProjectWorkflow} from "./mixtape/NewProjectWorkflow";

export const App = () => {
    const me = getSession();

    return <BrowserRouter>
        <p>Logged in as {me.Key} | {me.DiscordID} | {me.Roles ? me.Roles.join(", ") : UserRoles.Anonymous}</p>
        <Switch>
            <PrivateRoute
                path="/sessions"
                requireRole={UserRoles.ExecutiveDirector}>
                <Sessions/>
            </PrivateRoute>

            <PrivateRoute
                path="/mixtape/NewProjectWorkflow"
                requireRole={UserRoles.ExecutiveDirector}>
                <NewProjectWorkflow/>
            </PrivateRoute>

            <PrivateRoute
                path="/mixtape/"
                requireRole={UserRoles.VerifiedMember}>
                <MemberDashboard/>
            </PrivateRoute>

            <Route path="/login/failure"><LoginFailure/></Route>
            <Route path="/login/success"><LoginSuccess/></Route>
            <Route path="/login/discord"><LoginDiscord/></Route>
            <Route path="/login/"><Login/></Route>
            <Route path="/about/"><About/></Route>
            <Route path="/contact/"><Contact/></Route>
            <Route path="/"><Home/></Route>
        </Switch>
    </BrowserRouter>;
};

const PrivateRoute = (props: {
    path: string
    requireRole: string;
    children: JSX.Element;
}) => {
    const me = getSession();
    const authorized = me.Roles && me.Roles.includes(props.requireRole);
    const children = () => {
        switch (true) {
            case authorized:
                return props.children;
            case sessionIsAnonymous(me):
                return <Redirect to={"/login"}/>;
            default:
                return <Redirect to={"/401.html"}/>;
        }
    };
    return <Route path={props.path} render={children}/>;
};

ReactDOM.render(<App/>, document.querySelector("#entrypoint"));

