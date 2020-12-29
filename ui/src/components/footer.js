import React from 'react'

export default function Footer(props) {
    return <footer className="footer">
        <div className="container mt-3 text-center">
            <SocialMediaRow/>
            <PolicyRow/>
            <TeamsRow roles={props.roles}/>
        </div>
    </footer>
}

function SocialMediaRow() {
    function SocialMediaLink(props) {
        return <a className="text-light" href={props.href}>{props.children}</a>
    }

    return <div className="row">
        <div className="col">
            <SocialMediaLink href="https://www.youtube.com/channel/UCeipEtsfjAA_8ATsd7SNAaQ">
                <i className="fab fa-youtube fa-2x"/>
            </SocialMediaLink>
            <SocialMediaLink href="https://www.facebook.com/groups/1080154885682377/">
                <i className="fab fa-facebook fa-2x"/>
            </SocialMediaLink>
            <SocialMediaLink href="https://vvgo.bandcamp.com/">
                <i className="fab fa-bandcamp fa-2x"/>
            </SocialMediaLink>
            <SocialMediaLink href="https://github.com/virtual-vgo/vvgo">
                <i className="fab fa-github fa-2x"/>
            </SocialMediaLink>
            <SocialMediaLink href="https://www.instagram.com/virtualvgo/">
                <i className="fab fa-instagram fa-2x"/>
            </SocialMediaLink>
            <SocialMediaLink href="https://twitter.com/virtualvgo">
                <i className="fab fa-twitter fa-2x"/>
            </SocialMediaLink>
            <SocialMediaLink href="https://discord.com/invite/9RVUJMQ">
                <i className="fab fa-discord fa-2x"/>
            </SocialMediaLink>
        </div>
    </div>
}

function PolicyRow() {
    return <div className="row">
        <div className="col">
            <a className="text-light text-lowercase" href="https://vvgo.org/privacy">privacy policy</a>|
            <a className="text-light" href="https://vvgo.org/cookie-policy">cookie policy</a>
        </div>
    </div>
}

function TeamsRow(props) {
    if (props.roles.includes("vvgo-teams")) {
        return <div className="row alert-warning text-muted">
            <div className="col">
                <div className="dropdown">
                    <button className="dropdown-toggle btn btn-sm" type="button" data-toggle="dropdown">
                        View With Roles
                    </button>
                    <div className="dropdown-menu">
                        <ChooseRolesForm roles={props.roles}/>
                    </div>
                </div>
            </div>
        </div>
    } else {
        return null
    }
}

function ChooseRolesForm(props) {
    function RoleCheckbox(props) {
        return <div className="form-check">
            <input type="checkbox" className="form-check-input" name="roles" value={props.role}/>
            <label className="form-check-label" htmlFor="role">{props.role}</label>
        </div>
    }

    return <form className="px-2">
        {props.roles.map(role => <RoleCheckbox key={role} role={role}/>)}
        <div className="form-check">
            <input type="checkbox" className="form-check-input" name="roles" value="anonymous"/>
            <label className="form-check-label" htmlFor="role">anonymous</label>
        </div>
        <button type="submit" className="btn-sm btn-secondary">Submit</button>
    </form>
}
