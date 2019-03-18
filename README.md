# Petrify

An alternative to the hundreds of static website generators out there.  
Create your website in your web framework of choice and petrify will build and deploy it as a static website.

## Motivation

I wanted to make a static website for a friend.
After wading through all the most popular static website generators out there, I gave up and wrote the thing in [flask](http://flask.pocoo.org/).
But building and deploying took a lot more time and effort than I thought it would, especially so that a non-tech-savvy windows-using friend could manage and deploy it on his own, which is what petrify tries to solve.

## Dependencies

None. You don't even need git installed to deploy to github. 

## Supported operating systems

Linux, OS X and Windows

## Run it

Grab the [latest binary release](https://github.com/maggisk/petrify/releases) and add it to your project.
Create a .petrify configuration file in the same directory as the petrify binary and add the following

    # url to your development server
    WebsiteURL = "http://localhost:port"

    # github repository for the static website. the repository has to exist before you can deploy
    DeployToGithub = "git@github.com:username/username.git"

Run the petrify binary (you can double click it on OS X and Windows) and a browser tab will open previewing a build of your website.
If you are happy with it, answer yes to the prompt to deploy.  

If you want to run it from the command line you can also tell it what to do with argument `./petrify build`, `./petrify preview` or `./petrify deploy`

## Super quick, what else is needed to deploy to github
* Create a new public repo on github.com named __yourusername.github.io__
* If you have configured ssh access keys for github use the ssh version of the url (git@github.com...)
* If you don't, use the https version. You will be prompted for github username and password every time you deploy
* If you use the https version and have two-factor authentication enabled, you will have to [create an access token](https://help.github.com/en/articles/creating-a-personal-access-token-for-the-command-line) and use that in stead of your actual password
* If you have a domain for your website, set the CNAME config to `www.yourwebsite.com` and read the [github quick start on setting up a custom domain](https://help.github.com/en/articles/quick-start-setting-up-a-custom-domain)

## Configuration

Petrify will look for a `.petrify` [TOML](https://github.com/toml-lang/toml) configuration file in the same directory as the petrify binary

**Config name** | **Description** | **Default value** | **Example**
--- | --- || --
**ServerURL** | **Required.** URL to your development webserver | "" | ServerURL = "http://localhost:5000"
**CWD** | Set current working directory. All other filesystem paths can be set relative to this one | "." | CWD = "/home/maggisk/git/maggisk-website"
**EntryPoints** | List of paths where crawler should start crawling | ["/"] | EntryPoints = ["/", "/sitemap.xml", "/robots.txt"]
**BuildDir** | If you want to persist the build directory (not delete it when petrify exits) set this to a directory that petrify can write to. If empty, petrify will create a temporary directory for the build | "" | BuildDir = "./build"
**StaticDirs** | Directories with static files. Static files are usually not all discoverable by the crawler (e.g. when only linked to from css). Each entry is a source and destination directory separated by a semicolon. Destination directory should either by an absolute path or relative to the build directory | [] | StaticDirs = ["./webapp/static-files:static"]
**PreviewBeforeDeploy** | Petrify will by default start a webserver and open a browser window to view the built site. Set this to false if you want to skip this step and go straight to deploying | true | PreviewBeforeDeploy = false
**ExtractLinks** | Petrify will inspect links in all src and href attributes it finds. If you create links in any other ways, e.g. using data-src attributes to lazy load images, set this so petrify knows it should extract those links. [tagname].[attribute] format where tagname and attribute are either * or name of tag/attribute | "" | ExtractLinks = "img.data-src *.data-link"
**Path404** | Path to 404 page if you have one | "" | Path404 = "/404.html"
**DeployToGithub** | url to github repo hosting the website | "" | DeployToGithub = "git@github.com:maggisk/maggisk.git"
**CNAME** | If you have a custom domain for your website, set it here. | "" | CNAME = "www.example.com"
**Verbose** | Debug mode to see what petrify is doing | false | Verbose = true

## Things to keep in mind

* Static webservers do not support query parameters. All parameters to all pages have to be a part of the path. E.g. /articles/1/ and not /articles?id=1
* It's better to have your paths end with a slash. Static webservers will redirect /articles to /articles/ to find the index.html
* If you create an admin interface to manage your site, you don't even need to protect it with a password. As long as it is not linked to from any of the public pages, it will not become a part of the build

## TODOS
* Parse command line arguments that will have priority over .petrify config file
* Add an option to start the development webserver from the petrify process to make it more idiot proof. (a little tricky to be make sure you never leave orphan processes running)
* More deployment options: S3, gitlab (more?)
* Nicer errors on non-exceptional conditions (e.g. don't print stacktrace when config is missing)
