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
import {NotFound, AccessDenied, InternalOopsie} from "./components/error_page";
import Home from "./components/home";
import Helmet from "react-helmet";
import favicon from './favicons/favicon-2020-11-26-thomas.png'

ReactDOM.render(
    <BrowserRouter>
        <Helmet>
            <link rel="icon" href={favicon} sizes="32x32" type="image/png"/>
        </Helmet>
        <Switch>
            <Route exact path="/"><Nav><Home/></Nav></Route>
            <Route path="/about"><Nav><About/></Nav></Route>
            <Route path="/parts"><Nav><Parts/></Nav></Route>
            <Route path="/401.html"><AccessDenied/></Route>
            <Route path="/404.html"><NotFound/></Route>
            <Route path="/500.html"><InternalOopsie/></Route>
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
