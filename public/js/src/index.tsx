import React = require("react");
import ReactDOM = require("react-dom");
import {BrowserRouter as Router, Redirect, Route, Switch} from "react-router-dom";
import {About} from "./about";
import {getSession} from "./auth";
import {Roles, sessionIsAnonymous} from "./datasets";
import {Home} from "./home";
import {Login, LoginDiscord, LoginFailure, LoginSuccess} from "./login";
import {MemberDashboard} from "./mixtape/MemberDashboard";
import {NewProjectWorkflow} from "./mixtape/NewProjectWorkflow";
import {Sessions} from "./sessions";

const Routes = () => {
    const me = getSession();

    return <Router>
        <p>Logged in as {me.Key} | {me.DiscordID} | {me.Roles ? me.Roles.join(", ") : Roles.Anonymous}</p>
        <Switch>
            {/*<PrivateRoute*/}
            {/*    path="/sessions"*/}
            {/*    requireRole={Roles.ExecutiveDirector}>*/}
            {/*    <Sessions/>*/}
            {/*</PrivateRoute>*/}
            <Route path={"/sessions/"}><Sessions/></Route>

            <PrivateRoute
                path="/mixtape/NewProjectWorkflow"
                requireRole={Roles.ExecutiveDirector}>
                <NewProjectWorkflow/>
            </PrivateRoute>

            <PrivateRoute
                path="/mixtape/"
                requireRole={Roles.Member}>
                <MemberDashboard/>
            </PrivateRoute>

            <Route path="/login/failure"><LoginFailure/></Route>
            <Route path="/login/success"><LoginSuccess/></Route>
            <Route path="/login/discord"><LoginDiscord/></Route>
            <Route path="/login"><Login/></Route>
            <Route path="/about"><About/></Route>
            <Route path="/"><Home/></Route>
        </Switch>
    </Router>;
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

ReactDOM.render(<Routes/>, document.querySelector("#entrypoint"));

