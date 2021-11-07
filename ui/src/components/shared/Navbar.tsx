import {CSSProperties} from "react";
import {Dropdown, Nav, Navbar as BootstrapNavbar} from "react-bootstrap";
import {Link, NavLink as RouterNavLink} from "react-router-dom";
import {getSession} from "../../auth";
import {links} from "../../data/links";
import {sessionIsAnonymous, UserRole} from "../../datasets";
import {Favicon} from "./Favicon";

const navbarStyle: CSSProperties = {
    marginTop: "15px",
    backgroundColor: "rgba(71, 54, 80, 0.57)",
};

export const Navbar = () => {
    const me = getSession();
    const NavLink = (props: {
        to: string,
        children: string | JSX.Element | (JSX.Element | string)[]
    }) => <RouterNavLink
        to={props.to}
        activeClassName="bg-vvgo-purple nav-link"
        className="nav-link text-light">
        {props.children}
    </RouterNavLink>;

    const MemberNavLink = (props: {
        to: string,
        children: string | (JSX.Element | string)[]
    }) => (me.Roles && me.Roles.includes(UserRole.VerifiedMember)) ?
        <NavLink {...props}>{props.children}</NavLink> : <div/>;

    const PrivateNavLink = (props: {
        to: string,
        requireRole: UserRole
        children: string | (JSX.Element | string)[]
    }) => (me.Roles && me.Roles.includes(props.requireRole)) ?
        <RouterNavLink
            to={props.to}
            activeClassName="alert-warning text-dark nav-link"
            className="text-warning nav-link">
            {props.children}
        </RouterNavLink> : <div/>;

    return <BootstrapNavbar expand="md" className="fa-border mb-4" style={navbarStyle}>
        <Link className="nav-link text-light navbar-brand" to="/">
            <Favicon/>
        </Link>
        <BootstrapNavbar.Toggle/>
        <BootstrapNavbar.Collapse>
            <BootstrapNavbar.Collapse>
                <Nav className="me-auto">
                    <MemberNavLink to="/parts/">Parts</MemberNavLink>
                    <MemberNavLink to="/mixtape/">Wintry Mix</MemberNavLink>
                    <NavLink to="/projects/">Projects</NavLink>
                    <NavLink to="/about/">About</NavLink>
                    <NavLink to="/contact/">Contact</NavLink>
                    <Dropdown as={Nav.Item}>
                        <Dropdown.Toggle className="text-light" as={Nav.Link}>
                            Store
                        </Dropdown.Toggle>
                        <Dropdown.Menu>
                            <Dropdown.Item href={links.BandCamp}>Music</Dropdown.Item>
                            <Dropdown.Item href="/store">Merch</Dropdown.Item>
                        </Dropdown.Menu>
                    </Dropdown>
                    <PrivateNavLink
                        to="/credits-maker"
                        requireRole={UserRole.ProductionTeam}>
                        Credits Maker <i className="fas fa-lock"/>
                    </PrivateNavLink>
                </Nav>
                <Nav>{sessionIsAnonymous(me) ?
                    <NavLink to="/login">Login</NavLink> :
                    <NavLink to="/logout">Logout</NavLink>}
                </Nav>
            </BootstrapNavbar.Collapse>
        </BootstrapNavbar.Collapse>
    </BootstrapNavbar>;
};
