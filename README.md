# Togo4: On Street/Service Project
    Telegram bot (By webhook) version of TogoFor. for managing my todos, in order to make me go for them.
    With many extra features and Memory/Performance & Coding optimization.
    This bot application is running on Vercel as a Serverless Function bot.
# Language: GO
# Link
    running on https://t.me/togo4bot
# Notes:
* Here command/param seperator is 2 SPACES (because telegram doesnt have a specific tab character)
* More than 2 spaces is still part of the arguments; Separator is Exactly 2 spaces; nothing more of less!
* Set these Environmental Variables for start:
TELEGRAM_TOKEN=token
POSTGRES_URL=postgres connection string

# Markup Keyboard
   Comparing to togo4 console app, this one has many extra features icluding a Reply Markup keyboard and Inline keyboards on many section,
   Making it easyier to interact with the app.
# Commands
# +: New Togo:
=> ... +   title   [=  weight]    [+p   progress_till_now]   [:   description]    [+x | -x]   [@  start_date_as_how_many_days_from_now    start_time_as_hh:mm]    [NEXT_COMMAND]

*   Flags order are optional, and Flags and their params must be seperated by 2 SPACES.
*   weight value can also be set by +w flag
*   description value can also be set by +d flag
# #: Show Togos
=> ...   #   [NEXT_COMMAND]
    by default shows today's togos
=> ...   #   -a   [NEXT_COMMAND]
    Show all togos on any day
# %: Progress Made:
=> ...   %   [NEXT_COMMAND]
    Calculate the progress been made (by default for Today)
=> ...   %   -a [NEXT_COMMAND]
    Calculate the progress been made considering all togos on any day

# $: Get and Update a togo of today
=> ... $   id   [NEXT_COMMAND]
*   this will get and show a togo (just in today)
=> ... $   id   [=  weight]    [+p   progress_till_now]   [:   description]    [+x | -x]   [@  start_date_as_how_many_days_from_now    start_time_as_hh:mm]    [NEXT_COMMAND]

# $   -a: Get and Update ANY togo
=> ... $   -a   id   [=  weight]    [+p   progress_till_now]   [:   description]    [+x | -x]   [@  start_date_as_how_many_days_from_now    start_time_as_hh:mm]

# Other Notes:
*   ... means that these cammands can also be used after previous command in the same line.
*   Each line can contain multiple command, as many as you want. Like:

=>   +   new_togo    @   1   10:00   +p  85  #  +   next_togo   +x  #   %

*   Extra:
=>        +x: its an extra Togo. its not mandatory but has extra points doing it.
=>        -x: not extra (default)
*   all params between [] are optional.


# Remember:
*   The flag list separator is 2 SPACES. space character will be evaluated as a part of the current flag's param. do not be mistaken.
*   in 'add new togo' syntax, all flags are optional except for the title, meaning that you can simply add new togos even with specifying the title only such as:
=>  +   new togo here
*   use -a flag for % and # commands to expand the togos range to ALL Times.
    
# P.S.:
   Street/Service Project, means that this one is coded while walking streets or while doing service!
