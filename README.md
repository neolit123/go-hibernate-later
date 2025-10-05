# go-hibernate-later

## Description

A small Windows taskbar application, written in Go, for putting Windows into
hibernate after a timeout of inactivity.

## Why this exists?

There are some weird bugs around Windows 11's sleep and hibernate modes
on some setups.

Sleep can put the machine fans into burst mode for some reason.
Hibernate is disabled by default. Enabling it in the start menu and setting
a 'hibernate after' either via the GUI or via 'powercfg' still does not
trigger it.

## How it works

Every few seconds the application checks if there is no user input activity.
It also checks if there are no applications or drivers having 'requests' for the system
to be awake. If there is no input and no requests after the timeout, hibernate is triggered.

To get the active requests it calls 'powercfg /requests' which requires administrator.

## Installing

### Building the executable

- Install a Go compiler.
- Install go-winres.
- Download the source code of this application in a directory.
- Open a terminal and 'cd' to the directory.
- Call 'build.cmd'.

## Running the application on system startup as administrator

- Open Windows Task Scheduler.
- Create a new task.
- Enter a name.
- Select 'Run with highest privileges'.
- Under 'Triggers' select 'At log on'.
- Under 'Actions' select 'Start a program' and enter the executable path.
- Put '--timeout MIN' in the 'Add arguments' field, where MIN is the minutes
of timeout.
- Right click and run the task to test it.

## Usage

- Start the application as administrator.
- To exit the application right click the taskbar icon and select 'Exit'.
- Hold your mouse over the icon to see some information like the timeout, idle time
and requests.

## License

Apache 2.0. See the [LICENSE](./LICENSE) file.

## Icons and art

Authored by me (neolit123) using Krita.
The art license is the same as the project [LICENSE](./LICENSE).
