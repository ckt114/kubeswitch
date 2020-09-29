set __fish_kubeswitch_commands context ctx namespace ns

function __fish_kubeswitch_needs_command
  for i in (commandline -opc)
    if contains -- $i $__fish_kubeswitch_commands
      echo "$i"
      return 1
    end
  end
  return 0
end

function __fish_kubeswitch_using_command
  set -l cmd (__fish_kubeswitch_needs_command)
  test -z "$cmd"
  and return 1

  contains -- $cmd $argv
  and echo "$cmd"
  and return 0

  return 1
end


complete -f -c kubeswitch -f -n '__fish_kubeswitch_needs_command' -a "$__fish_kubeswitch_commands"

for subcmd in "context ctx"
  complete -f -c kubeswitch -f -n "__fish_kubeswitch_using_command $subcmd" -a '(kubeswitch ctx -P)'
end

for subcmd in "namespace ns"
  complete -f -c kubeswitch -f -n "__fish_kubeswitch_using_command $subcmd" -a '(kubeswitch ns -P)'
end

