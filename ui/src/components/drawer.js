import React, {useState} from 'react';
import clsx from 'clsx';
import {makeStyles, useTheme} from '@material-ui/core/styles';
import Drawer from '@material-ui/core/Drawer';
import CssBaseline from '@material-ui/core/CssBaseline';
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import List from '@material-ui/core/List';
import Typography from '@material-ui/core/Typography';
import Divider from '@material-ui/core/Divider';
import IconButton from '@material-ui/core/IconButton';
import MenuIcon from '@material-ui/icons/Menu';
import ChevronLeftIcon from '@material-ui/icons/ChevronLeft';
import ChevronRightIcon from '@material-ui/icons/ChevronRight';
import ListItem from '@material-ui/core/ListItem';
import ListItemText from '@material-ui/core/ListItemText';
import {Link as RouteLink, NavLink as RouterLink} from "react-router-dom";
import {ExpandLess, ExpandMore} from "@material-ui/icons";
import Collapse from "@material-ui/core/Collapse";

const drawerWidth = 300;

const useStyles = makeStyles((theme) => ({
    root: {display: 'flex'},
    appBar: {
        transition: theme.transitions.create(['margin', 'width'], {
            easing: theme.transitions.easing.sharp,
            duration: theme.transitions.duration.leavingScreen,
        }),
    },
    appBarShift: {
        width: `calc(100% - ${drawerWidth}px)`,
        marginLeft: drawerWidth,
        transition: theme.transitions.create(['margin', 'width'], {
            easing: theme.transitions.easing.easeOut,
            duration: theme.transitions.duration.enteringScreen,
        }),
    },
    menuButton: {marginRight: theme.spacing(2)},
    hide: {display: 'none'},
    drawer: {width: drawerWidth, flexShrink: 0},
    drawerPaper: {width: drawerWidth},
    drawerHeader: {
        display: 'flex',
        alignItems: 'center',
        padding: theme.spacing(0, 1),
        // necessary for content to be below app bar
        ...theme.mixins.toolbar,
        justifyContent: 'flex-end',
    },
    content: {
        flexGrow: 1,
        padding: theme.spacing(3),
        transition: theme.transitions.create('margin', {
            easing: theme.transitions.easing.sharp,
            duration: theme.transitions.duration.leavingScreen,
        }),
        marginLeft: -drawerWidth,
    },
    contentShift: {
        transition: theme.transitions.create('margin', {
            easing: theme.transitions.easing.easeOut,
            duration: theme.transitions.duration.enteringScreen,
        }),
        marginLeft: 0,
    },
    nested: {
        paddingLeft: theme.spacing(4),
    },
}));

export default function AppDrawer(props) {
    const classes = useStyles();
    const theme = useTheme();

    // drawer
    const [drawerState, setDrawerState] = useState(true);
    const openDrawer = () => setDrawerState(true);
    const closeDrawer = () => setDrawerState(false);

    return (
        <div className={classes.root}>
            <CssBaseline/>
            <AppBar position="fixed" className={clsx(classes.appBar, {[classes.appBarShift]: drawerState})}>
                <Toolbar>
                    <IconButton aria-label="open drawer" color="inherit" edge="start" onClick={openDrawer}
                                className={clsx(classes.menuButton, drawerState && classes.hide)}>
                        <MenuIcon/>
                    </IconButton>
                    <Typography variant="h6" noWrap>{props.appTitle}</Typography>
                </Toolbar>
            </AppBar>
            <Drawer className={classes.drawer} variant="persistent" anchor="left"
                    open={drawerState} classes={{paper: classes.drawerPaper,}}>
                <div className={classes.drawerHeader}>
                    <Typography>Virtual VGO</Typography>
                    <IconButton onClick={closeDrawer}>
                        {theme.direction === 'ltr' ? <ChevronLeftIcon/> : <ChevronRightIcon/>}
                    </IconButton>
                </div>
                <Divider/>
                <List>
                    <ListItem divider><ListItemText primary="Open Projects"/></ListItem>
                    <OpenProjects roles={props.uiRoles.data} projects={props.projects} parts={props.parts}/>
                    <Divider/>
                    <ListItem divider><ListItemText primary="Releases"/></ListItem>
                    <Releases projects={props.projects}/>
                    <Divider/>
                    <MyListItem to={"/about"}>About</MyListItem>
                    <TeamsListItem roles={props.uiRoles.data} to={"/credits-maker"}>Credits Maker</TeamsListItem>
                </List>
            </Drawer>
            <main className={clsx(classes.content, {[classes.contentShift]: drawerState})}>
                <div className={classes.drawerHeader}/>
                {props.children}
            </main>
        </div>
    );
}

function MyListItem(props) {
    return <ListItem divider button component={RouterLink} to={props.to}>
        <ListItemText primary={props.children}/>
    </ListItem>
}

function TeamsListItem(props) {
    if (props.roles.includes("vvgo-teams")) {
        return <ListItem divider button component={RouterLink} to={props.to}>
            <ListItemText primary={props.children} color='secondary'/>
        </ListItem>
    } else {
        return null
    }
}

function Releases(props) {
    const projects = props.projects.filter(project => (project.VideoReleased === true))
    projects.sort((a, b) => b.Name.localeCompare(a.Name))
    return projects
        .map(project =>
            <ListItem key={project.Name} button component={RouteLink} to={`/releases/${project.Name}`}>
                <ListItemText primary={project.Title}/>
            </ListItem>
        )
}

function OpenProjects(props) {
    return props.projects
        .filter(project => (project.PartsReleased === true && project.PartsArchived === false))
        .map(project => <PartListing key={project.Name} project={project} parts={props.parts}/>)
}

function PartListing(props) {
    const classes = useStyles();
    const [open, setOpen] = useState(false);
    const handleClick = () => setOpen(!open);

    const parts = props.parts
        .filter(part => (part.Project === props.project.Name))

    return <div>
        <ListItem button onClick={handleClick}>
            <ListItemText primary={props.project.Title}/>{open ? <ExpandLess/> : <ExpandMore/>}</ListItem>
        <Collapse in={open} timeout="auto" unmountOnExit>
            <List component="div" disablePadding>
                {parts.map(part =>
                    <ListItem key={part.PartName} button className={classes.nested}
                              component={RouteLink} to={`/parts/${props.project.Name}/${part.PartName}`}>
                        {part.PartName}
                    </ListItem>
                )}
            </List>
        </Collapse>
    </div>
}
