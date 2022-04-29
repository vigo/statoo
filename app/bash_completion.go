package app

var bashCompletion = `__statoo_comp()
{
    local cur next
    cur="${COMP_WORDS[COMP_CWORD]}"
    opts="-a -auth -f -find -header -h -help -j -json -t -timeout -s -skip -verbose -version"
    COMPREPLY=( $(compgen -W "${opts}" -- "${cur}") )
}
complete -F __statoo_comp statoo`
