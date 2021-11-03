import React = require("react");
import {BrowserRouter, Redirect, Route, Switch} from "react-router-dom";
import {getSession} from "../auth";
import {sessionIsAnonymous, UserRole} from "../datasets";
import {About} from "./About";
import {Contact} from "./Contact";
import {CreditsMaker} from "./CreditsMaker";
import {AccessDenied} from "./errors/AccessDenied";
import {InternalOopsie} from "./errors/InternalOopsie";
import {NotFound} from "./errors/NotFound";
import {Home} from "./Home";
import {Login} from "./login/Login";
import {LoginDiscord} from "./login/LoginDiscord";
import {LoginFailure} from "./login/LoginFailure";
import {Logout} from "./login/Logout";
import {MemberDashboard} from "./mixtape/MemberDashboard";
import {NewProjectFormResponses} from "./mixtape/NewProjectFormResponses";
import {NewProjectWorkflow} from "./mixtape/NewProjectWorkflow";
import {Parts} from "./Parts";
import {Projects} from "./Projects";
import {Sessions} from "./Sessions";
import {Members as MemberStats} from "./stats/Members";

export const App = () => {
    document.documentElement.lang = "en";
    return <BrowserRouter>
        <Switch>
            <PrivateRoute path="/credits-maker/" role={UserRole.ProductionTeam}><CreditsMaker/></PrivateRoute>
            <PrivateRoute path="/stats/members" role={UserRole.VerifiedMember}><MemberStats/></PrivateRoute>
            <PrivateRoute path="/parts/" role={UserRole.VerifiedMember}><Parts/></PrivateRoute>
            <PrivateRoute
                path="/mixtape/NewProjectWorkflow/"
                role={UserRole.ExecutiveDirector}>
                <NewProjectWorkflow/>
            </PrivateRoute>
            <PrivateRoute
                path="/mixtape/NewProjectFormResponses/"
                role={UserRole.ExecutiveDirector}>
                <NewProjectFormResponses/>
            </PrivateRoute>
            <PrivateRoute path="/mixtape/" role={UserRole.VerifiedMember}><MemberDashboard/></PrivateRoute>
            <PrivateRoute path="/sessions/" role={UserRole.ExecutiveDirector}><Sessions/></PrivateRoute>

            <Route path="/login/failure/"><LoginFailure/></Route>
            <Route path="/login/discord/"><LoginDiscord/></Route>
            <Route path="/logout/"><Logout/></Route>
            <Route path="/login/"><Login/></Route>
            <Route path="/projects/"><Projects/></Route>
            <Route path="/about/"><About/></Route>
            <Route path="/contact/"><Contact/></Route>
            <Route exact path="/401.html"><AccessDenied/></Route>
            <Route exact path="/404.html"><NotFound/></Route>
            <Route exact path="/500.html"><InternalOopsie/></Route>
            <Route exact path="/"><Home/></Route>
            <Route path="*"><NotFound/></Route>
        </Switch>
    </BrowserRouter>;
};

const PrivateRoute = (props: {
    path: string
    role: string;
    children: JSX.Element;
}) => {
    const me = getSession();
    const authorized = me.Roles && me.Roles.includes(props.role);
    const children = () => {
        switch (true) {
            case authorized:
                return props.children;
            case sessionIsAnonymous(me):
                return <Redirect to={"/login"}/>;
            default:
                return <AccessDenied/>;
        }
    };
    return <Route path={props.path} render={children}/>;
};
