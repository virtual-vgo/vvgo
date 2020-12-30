import React from 'react'
import {NavLink} from 'react-router-dom'
import Container from "@material-ui/core/Container";
import Link from "@material-ui/core/Link";

export default function Navbar(props) {
    return <Container>
        <nav className="top-nav navbar navbar-expand-md navbar-dark bg-dark-blue-transparent fa-border">
            <NavLink className="nav-link" to="/">
                <img src={props.favicon} alt="favicon"/>
            </NavLink>
            <button className="navbar-toggler" type="button" data-toggle="collapse" data-target="#navbarCollapse"
                    aria-controls="navbarCollapse" aria-expanded="false" aria-label="Toggle navigation">
                <span className="navbar-toggler-icon"/>
            </button>
            <div className="collapse navbar-collapse" id="navbarCollapse">
                <ul className="navbar-nav mr-auto">
                    <MemberNavItem to="/parts" roles={props.roles}>Parts</MemberNavItem>
                    <NavItem to="/projects">Projects</NavItem>
                    <NavItem to="/about">About</NavItem>
                    <TeamsNavItem to="/credits-maker" roles={props.roles}>Credits Maker</TeamsNavItem>
                </ul>
                <ul className="navbar-nav mr-2">
                    <LoginNavItem roles={props.roles}/>
                </ul>
            </div>
        </nav>
    </Container>
}

function NavItem(props) {
    return <li className="nav-item">
        <NavLink to={props.to} activeClassName="bg-vvgo-purple" className="nav-link">
            {props.children}
        </NavLink>
    </li>
}

function MemberNavItem(props) {
    if (props.roles.includes("vvgo-member")) {
        return <NavItem to={props.to}>{props.children}</NavItem>
    } else {
        return null
    }
}

function TeamsNavItem(props) {
    if (props.roles.includes("vvgo-teams")) {
        return <li className="nav-item">
            <NavLink to={props.to} activeClassName="alert-warning text-dark" className="nav-link text-warning">
                {props.children}
            </NavLink>
        </li>
    } else {
        return null
    }
}

function LoginNavItem(props) {
    function userLoggedIn(roles) {
        switch (roles.length) {
            case 0:
                return false
            case 1:
                return roles[0] !== "anonymous"
            default:
                return true
        }
    }

    if (userLoggedIn(props.roles)) {
        return <Link href="/logout">Logout</Link>
    } else {
        return <NavItem to="/login">Login</NavItem>
    }
}
