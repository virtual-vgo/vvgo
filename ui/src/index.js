import React, {useState} from 'react'
import ReactDOM from 'react-dom'
import 'bootstrap/dist/css/bootstrap.min.css'
import './css/theme.module.css'
import Footer from './components/footer'
import About from './components/about'
import {BrowserRouter, Route, Switch} from "react-router-dom";
import reportWebVitals from "./reportWebVitals";
import {AccessDenied, InternalOopsie, NotFound} from "./components/error_page";
import Home from "./components/home";
import 'bootstrap/dist/js/bootstrap.bundle.min'
import '@fortawesome/fontawesome-free/css/all.min.css'
import '@fortawesome/fontawesome-free/js/fontawesome.min.js'
import {useLeaders, useLoginRoles, useParts, useProjects} from "./components/hooks";
import DevTools from "./components/dev_tools";
import createMuiTheme from "@material-ui/core/styles/createMuiTheme";
import {ThemeProvider} from '@material-ui/core/styles';
import AppDrawer from "./components/drawer";
import Part from "./components/part";

const theme = createMuiTheme({
    typography: {fontFamily: 'Montserrat, sans-serif'},
    palette: {type: 'dark'}
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

    const [appTitle, setAppTitle] = useState('Virtual VGO')

    function Nav(props) {
        return <div>
            <AppDrawer uiRoles={uiRoles} projects={projects.data} parts={parts.data} appTitle={appTitle}>
                <DevTools uiRoles={uiRoles} apiRoles={apiRoles}/>
                {props.children}
                <Footer uiRoles={uiRoles} apiRoles={apiRoles}/>
            </AppDrawer>
        </div>
    }

    function PartRouter(props) {
        // index projects by project name
        const projectIndex = {}
        props.projects.forEach(project => projectIndex[project.Name] = project)

        return props.parts.map(part =>
            <Route key={part.PartName} path={`/parts/${part.Project}/${part.PartName}`}>
                <Part setAppTitle={setAppTitle} project={projectIndex[part.Project]} part={part}/>
            </Route>
        )
    }

    return <BrowserRouter>
        <Switch>
            <Route exact path="/401.html"><AccessDenied/></Route>
            <Route exact path="/404.html"><NotFound/></Route>
            <Route exact path="/500.html"><InternalOopsie/></Route>
            <Route path="/">
                <Nav>
                    <Switch>
                        <Route exact path="/"><Home projects={projects.data}/></Route>
                        <Route path="/about"><About leaders={leaders.data}/></Route>
                        <PartRouter parts={parts.data} projects={projects.data}/>
                    </Switch>
                </Nav>
            </Route>
        </Switch>
    </BrowserRouter>
}

// ref: https://bit.ly/CRA-vitals
reportWebVitals(console.log);
