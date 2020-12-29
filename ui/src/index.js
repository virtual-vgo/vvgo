import React from 'react'
import ReactDOM from 'react-dom'
import 'bootstrap/dist/css/bootstrap.min.css'
import './theme.css'
import Footer from './components/footer'
import About from './components/about'
import Parts from './components/parts'
import {BrowserRouter, Route, Switch} from "react-router-dom";
import Navbar from './components/navbar'
import reportWebVitals from "./reportWebVitals";
import {NotFound} from "./components/error_page";
import Home from "./components/home";

ReactDOM.render(
    <BrowserRouter>
        <Switch>
            <Route path="/about"><Nav><About/></Nav></Route>
            <Route path="/parts"><Nav><Parts/></Nav></Route>
            <Route exact path="/"><Nav><Home/></Nav></Route>
            <Route path="*"><NotFound/></Route>
        </Switch>
    </BrowserRouter>,
    document.getElementById('root')
)

function Nav(props) {
    return [<Navbar key="navbar"/>, props.children, <Footer key="footer"/>]
}

// ref: https://bit.ly/CRA-vitals
reportWebVitals(console.log);
