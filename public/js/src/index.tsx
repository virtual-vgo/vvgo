import React = require("react");
import ReactDOM = require("react-dom");
import {BrowserRouter as Router, Route, Switch} from "react-router-dom";
import {About} from "./about";
import {Home} from "./home";
import {Mixtape} from "./mixtape";
import {NewProjectWorkflow} from "./mixtape/NewProjectWorkflow";
import {Sessions} from "./sessions";


ReactDOM.render(<Router>
    <Switch>
        <Route path="/sessions"><Sessions/></Route>
        <Route path="/mixtape/NewProjectWorkflow"><NewProjectWorkflow/></Route>
        <Route path="/mixtape"><Mixtape/></Route>
        <Route path="/about"><About/></Route>
        <Route path="/"><Home/></Route>
    </Switch>
</Router>, document.querySelector("#entrypoint"));

