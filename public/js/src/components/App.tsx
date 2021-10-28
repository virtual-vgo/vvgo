import React = require("react");
import {BrowserRouter, Redirect, Route, Switch} from "react-router-dom";
import {getSession} from "../auth";
import {sessionIsAnonymous, UserRoles} from "../datasets";
import {About} from "./About";
import {Contact} from "./Contact";
import {Home} from "./Home";
import {Login} from "./login/Login";
import {LoginDiscord} from "./login/LoginDiscord";
import {LoginFailure} from "./login/LoginFailure";
import {LoginSuccess} from "./login/LoginSuccess";
import {Logout} from "./login/Logout";
import {MemberDashboard} from "./mixtape/MemberDashboard";
import {NewProjectWorkflow} from "./mixtape/NewProjectWorkflow";
import {Parts} from "./Parts";
import {Sessions} from "./Sessions";

export const App = () => {
    return <BrowserRouter>
        {JSON.stringify(getSession())}
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

            <PrivateRoute
                path="/parts/"
                requireRole={UserRoles.VerifiedMember}>
                <Parts/>
            </PrivateRoute>

            <Route path="/login/failure"><LoginFailure/></Route>
            <Route path="/login/success"><LoginSuccess/></Route>
            <Route path="/login/discord"><LoginDiscord/></Route>
            <Route path="/logout"><Logout/></Route>
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
