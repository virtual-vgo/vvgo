import React from 'react'

export default function Footer() {
    return <footer className="footer">
        <div className="container mt-3">
            <SocialMediaRow/>
            <PolicyRow/>
        </div>
    </footer>
}

function SocialMediaRow() {
    function SocialMediaLink(props) {
        return <a className="text-light" href={props.href}>{props.children}</a>
    }

    return <div className="row text-center">
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
    return <div className="row text-center">
        <div className="col">
            <a className="text-light text-lowercase" href="https://vvgo.org/privacy">privacy policy</a>|
            <a className="text-light" href="https://vvgo.org/cookie-policy">cookie policy</a>
        </div>
    </div>
}
