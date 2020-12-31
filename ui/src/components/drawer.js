import React, {useState} from 'react';
import clsx from 'clsx';
import Drawer from '@material-ui/core/Drawer';
import CssBaseline from '@material-ui/core/CssBaseline';
import List from '@material-ui/core/List';
import Typography from '@material-ui/core/Typography';
import Divider from '@material-ui/core/Divider';
import IconButton from '@material-ui/core/IconButton';
import ChevronLeftIcon from '@material-ui/icons/ChevronLeft';
import ListItem from '@material-ui/core/ListItem';
import ListItemText from '@material-ui/core/ListItemText';
import {Link as RouteLink} from "react-router-dom";
import {ExpandLess, ExpandMore} from "@material-ui/icons";
import Collapse from "@material-ui/core/Collapse";
import {useVVGOStyles} from "./styles";
import SearchIcon from '@material-ui/icons/Search';
import InputBase from '@material-ui/core/InputBase';
import Footer from "./footer";

const useStyles = useVVGOStyles

export default function AppDrawer(props) {
    const classes = useStyles();

    const [searchState, setSearchState] = useState('')
    const parts = props.parts.filter(part =>
        `${part.PartName}`
            .toLowerCase()
            .replaceAll('♭', 'b')
            .includes(searchState)
    )

    return (
        <div className={classes.root}>
            <CssBaseline/>
            <Drawer className={classes.drawer} variant="persistent" anchor="left"
                    open={props.drawerState.isOpen} classes={{paper: classes.drawerPaper,}}>
                <div className={classes.drawerHeader}>
                    <Typography>Open Projects</Typography>
                    <IconButton onClick={props.drawerState.closeDrawer}>
                        <ChevronLeftIcon/>
                    </IconButton>
                </div>
                <Divider/>
                <List>
                    <ListItem><ListItemText primary='Open Projects'/></ListItem>
                    <Search setSearchState={setSearchState}/>
                    <Divider/>
                    <OpenProjects projects={props.projects} parts={parts}/>
                    <Divider/>
                </List>
                <div style={{height: '100%'}}/>
                <Footer/>
            </Drawer>
            <main className={clsx(classes.content, {[classes.contentShift]: props.drawerState.isOpen})}>
                <div className={classes.drawerHeader}/>
                {props.children}
            </main>
        </div>
    );
}

function Search(props) {
    const classes = useStyles();
    const updateSearch = (event) => {
        props.setSearchState(event.target.value)
        console.log("new search update", event.target.value)
    }

    return <ListItem className={classes.search}>
        <div className={classes.searchIcon}>
            <SearchIcon/>
        </div>
        <InputBase
            placeholder="Search…"
            classes={{
                root: classes.searchInputRoot,
                input: classes.searchInput,
            }}
            onChange={updateSearch}
            inputProps={{'aria-label': 'search'}}
        />
    </ListItem>
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
                    <ListItem key={`${part.project} - ${part.PartName}`} button className={classes.nestedList}
                              component={RouteLink} to={`/parts/${props.project.Name}/${part.PartName}`}>
                        {part.PartName}
                    </ListItem>
                )}
            </List>
        </Collapse>
    </div>
}
