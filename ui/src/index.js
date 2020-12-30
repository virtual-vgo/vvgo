import React from 'react'
import ReactDOM from 'react-dom'
import 'bootstrap/dist/css/bootstrap.min.css'
import './theme.css'
import Footer from './components/footer'
import About from './components/about'
import Parts from './components/parts'
import {BrowserRouter, Route, Switch} from "react-router-dom";
import reportWebVitals from "./reportWebVitals";
import {AccessDenied, InternalOopsie, NotFound} from "./components/error_page";
import Home from "./components/home";
import Helmet from "react-helmet";
import favicon from './favicons/favicon-2020-11-26-thomas.png'
import 'bootstrap/dist/js/bootstrap.bundle.min'
import '@fortawesome/fontawesome-free/css/all.min.css'
import '@fortawesome/fontawesome-free/js/fontawesome.min.js'
import {useLeaders, useLoginRoles, useParts, useProjects} from "./components/hooks";
import DevTools from "./components/dev_tools";
import createMuiTheme from "@material-ui/core/styles/createMuiTheme";
import {ThemeProvider} from '@material-ui/core/styles';
import AppDrawer from "./components/drawer";

const theme = createMuiTheme({
    typography: {
        fontFamily: 'Montserrat, sans-serif'
    },
    palette: {
        type: 'dark'
    }
});

ReactDOM.render(
    <ThemeProvider theme={theme}>
        <App/>
    </ThemeProvider>, document.getElementById('root')
)

function App() {
    const apiRoles = useLoginRoles()
    const uiRoles = useLoginRoles()
    const parts = useParts()
    const projects = useProjects()
    const leaders = useLeaders()

    function Nav(props) {
        return <div>
            <AppDrawer uiRoles={uiRoles} key={"drawer"}>
                <DevTools uiRoles={uiRoles} apiRoles={apiRoles}/>
                {props.children}
                <Footer uiRoles={uiRoles} apiRoles={apiRoles}/>
            </AppDrawer>
        </div>
    }

    return <BrowserRouter>
        <Helmet>
            <link rel="icon" href={favicon} sizes="32x32" type="image/png"/>
        </Helmet>
        <Switch>
            <Route exact path="/"><Nav><Home projects={projects.data}/></Nav></Route>
            <Route path="/about"><Nav><About leaders={leaders.data}/></Nav></Route>
            <Route path="/parts"><Nav><Parts parts={parts.data} projects={projects.data}/></Nav></Route>
            <Route path="/401.html"><AccessDenied/></Route>
            <Route path="/404.html"><NotFound/></Route>
            <Route path="/500.html"><InternalOopsie/></Route>
            <Route path="*"><NotFound/></Route>
        </Switch>
    </BrowserRouter>
}

// ref: https://bit.ly/CRA-vitals
reportWebVitals(console.log);
