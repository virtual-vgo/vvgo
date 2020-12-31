import React from "react"
import {Link} from "react-router-dom"
import styles from '../css/error_page.module.css'
import notFoundGif from '../images/404.gif'
import accessDeniedGif from '../images/401.gif'
import internalOopsieGif from '../images/500.gif'

export function NotFound() {
    document.title = '404 Not Found'
    return <div>
        <img className={styles.errorImg} src={notFoundGif} alt="404 Not Found"/>
        <div className={styles.helpMe}>
            <Link to="/" className={styles.helpMe}>Click here to return to safety.</Link>
        </div>
    </div>
}

export function AccessDenied() {
    document.title = '401 Access Denied'
    return <div>
        <img className={styles.errorImg} src={accessDeniedGif} alt="404 Not Found"/>
        <div className={styles.helpMe}>
            <Link to="/" className={styles.helpMe}>Click here to return to safety.</Link>
        </div>
    </div>
}

export function InternalOopsie() {
    document.title = '500 Internal Oopsie'
    return <div>
        <img className={styles.errorImg} src={internalOopsieGif} alt="404 Not Found"/>
        <div className={styles.helpMe}>
            <Link to="/" className={styles.helpMe}>Click here to return to safety.</Link>
        </div>
    </div>
}
