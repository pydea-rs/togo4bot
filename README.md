# Togo4: 2nd Street Project
    Telegram bot (By webhook) version of TogoFor. for managing my todos, in order to make me go for them. 
# Language: GO

# Notes:
Set these Environmental Variables for start:
TELEGRAM_TOKEN=token
POSTGRES_URL=postgres connection string

# Commands
# +: New Togo:
=> ... +   title   [=  weight]    [+p   progress_till_now]   [:   description]    [+x | -x]   [@  start_date_as_how_many_days_from_now    start_time_as_hh:mm]    [NEXT_COMMAND]

*   Flags order are optional, and Flags and their params must be seperated by TABS.
*   weight value can also be set by +w flag
*   description value can also be set by +d flag
# #: Show Togos
=> ...	#	[NEXT_COMMAND]
    by default shows today's togos
# %: Progress Made:
=> ...	%	[NEXT_COMMAND]
    Calculate the progress been made (by default for Today)

# $: Get and Update a togo of today
=> ... $	id	[NEXT_COMMAND]
*	this will get and show a togo (just in today)
=> ... $	id	[=  weight]    [+p   progress_till_now]   [:   description]    [+x | -x]   [@  start_date_as_how_many_days_from_now    start_time_as_hh:mm]    [NEXT_COMMAND]
# ><    Exit

*   ... means that these cammands can also be used after previous command in the same line.
*   Each line can contain multiple command, as many as you want. Like:

=>   +   new_togo    @   1   10:00   +p  85  #  +   next_togo   +x  #   %

*   Extra:
=>        +x: its an extra Togo. its not mandatory but has extra points doing it.
=>        -x: not extra (default)
*   all params between [] are optional.

# Remember:
*   The flag list separator is 3 spaces. space character will be evaluated as a part f the current flag's param. do not be mistaken.
*   in add new togo syntax, all flags are optional except for the title, meaning that you can simply add new togos even with specifying the title only such as:
=>  +   new togo here
*   use -a flag for % and # commands to expand the togos range to ALL Times.
    
# P.S.:
	Street Project, means that this one is also coded while walking the streets.
