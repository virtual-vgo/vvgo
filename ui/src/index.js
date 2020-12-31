import React from 'react'
import ReactDOM from 'react-dom'
import Footer from './components/footer'
import {BrowserRouter, Route, Switch} from "react-router-dom";
import reportWebVitals from "./reportWebVitals";
import '@fortawesome/fontawesome-free/css/all.min.css'
import '@fortawesome/fontawesome-free/js/fontawesome.min.js'
import {useDrawerState, useLoginRoles, useParts, useProjects} from "./components/hooks";
import createMuiTheme from "@material-ui/core/styles/createMuiTheme";
import {ThemeProvider} from '@material-ui/core/styles';
import AppDrawer from "./components/drawer";
import Part from "./components/part";
import {NotFound} from "./components/error_page";
import {Container} from "@material-ui/core";
import {YoutubeIframe} from "./components/utils";
import VVGOAppBar from "./components/app_bar";

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
    const drawerState = useDrawerState(true)

    function Nav(props) {
        return <div>
            <AppDrawer uiRoles={uiRoles} projects={projects.data} parts={parts.data} drawerState={drawerState}>
                {props.children}
            </AppDrawer>
        </div>
    }

    function PartRoutes(props) {
        // index projects by project name
        const projectIndex = {}
        props.projects.forEach(project => projectIndex[project.Name] = project)
        return props.parts.map(part =>
            <Route key={`${part.Project}-${part.PartName}`} path={`/parts/${part.Project}/${part.PartName}`}>
                <Part drawerState={drawerState} project={projectIndex[part.Project]} part={part}/>
            </Route>
        )
    }

    return <BrowserRouter>
        <Nav>
            <Switch>
                <Route exact path='/'>
                    <VVGOAppBar drawerState={drawerState} title='Parts Browser'/>
                    <Container>
                        <YoutubeIframe src='https://www.youtube.com/embed/VgqtZ30bMgM'/>
                    </Container>
                </Route>
                <PartRoutes parts={parts.data} projects={projects.data}/>
                <Route path="*"><NotFound/></Route>
            </Switch>
        </Nav>
    </BrowserRouter>
}

// ref: https://bit.ly/CRA-vitals
reportWebVitals(console.log);
